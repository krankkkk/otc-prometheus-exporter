package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"unicode"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/internal"
)

// aomHostDimensions are AOM dimension names that identify the host itself.
// These are already represented by resource_id/resource_name and are skipped
// when building Prometheus labels to avoid redundancy.
var aomHostDimensions = map[string]bool{
	"hostID":      true,
	"hostName":    true,
	"nameSpace":   true,
	"clusterId":   true,
	"clusterName": true,
	"nodeName":    true,
	"nodeIP":      true,
}

type aomRequest struct {
	Metrics    []aomMetricQuery `json:"metrics"`
	Period     int              `json:"period"`
	Timerange  string           `json:"timerange"`
	Statistics []string         `json:"statistics"`
}

type aomMetricQuery struct {
	Namespace  string         `json:"namespace"`
	MetricName string         `json:"metricName"`
	Dimensions []aomDimension `json:"dimensions"`
}

type aomDimension struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type aomResponse struct {
	ErrorCode    string            `json:"errorCode"`
	ErrorMessage string            `json:"errorMessage"`
	Metrics      []aomMetricResult `json:"metrics"`
}

type aomMetricResult struct {
	Metric     aomMetricInfo  `json:"metric"`
	Datapoints []aomDatapoint `json:"dataPoints"`
}

type aomMetricInfo struct {
	Namespace  string         `json:"namespace"`
	MetricName string         `json:"metricName"`
	Dimensions []aomDimension `json:"dimensions"`
}

type aomDatapoint struct {
	Timestamp  int64          `json:"timestamp"`
	Unit       string         `json:"unit"`
	Statistics []aomStatistic `json:"statistics"`
}

type aomStatistic struct {
	Statistic string  `json:"statistic"`
	Value     float64 `json:"value"`
}

// aomListMetricsResponse is the response from POST /v1/{project_id}/ams/metrics.
type aomListMetricsResponse struct {
	ErrorCode    string          `json:"errorCode"`
	ErrorMessage string          `json:"errorMessage"`
	Metrics      []aomMetricInfo `json:"metrics"`
}

// aomSuccessCode is the error code AOM returns on success.
const aomSuccessCode = "SVCSTG_AMS_2000000"

// checkAOMError returns an error if the AOM response contains an application-level
// error code. AOM can return HTTP 200 with a non-success errorCode.
func checkAOMError(code, message string) error {
	if code != "" && code != aomSuccessCode {
		return fmt.Errorf("aom error %s: %s", code, message)
	}
	return nil
}

// discoverAOMMetrics uses the list-metrics API to find all metric+dimension
// combinations for a given hostID. This is necessary because per-device metrics
// (disk, network, filesystem) have extra dimensions (diskDevice, netDevice,
// mountPoint) that must be included in the data query to get per-device results.
func discoverAOMMetrics(ctx context.Context, aomClient *golangsdk.ServiceClient, hostID string) ([]aomMetricQuery, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	body := map[string]interface{}{
		"metricItems": []map[string]interface{}{
			{
				"namespace":  "PAAS.NODE",
				"dimensions": []map[string]string{{"name": "hostID", "value": hostID}},
			},
		},
	}

	var resp aomListMetricsResponse
	_, err := aomClient.Post(aomClient.ServiceURL("ams", "metrics")+"?limit=1000", body, &resp, &golangsdk.RequestOpts{
		OkCodes: []int{200},
	})
	if err != nil {
		return nil, fmt.Errorf("aom list metrics: %w", err)
	}
	if err := checkAOMError(resp.ErrorCode, resp.ErrorMessage); err != nil {
		return nil, err
	}

	var queries []aomMetricQuery
	for _, m := range resp.Metrics {
		dims := make([]aomDimension, len(m.Dimensions))
		copy(dims, m.Dimensions)
		queries = append(queries, aomMetricQuery{
			Namespace:  m.Namespace,
			MetricName: m.MetricName,
			Dimensions: dims,
		})
	}
	return queries, nil
}

// fetchAOMMetrics discovers all metric+dimension combinations for the given
// hostID, then queries data for each. The AOM API allows up to 20 metrics per
// data request, so queries are batched via SliceWindow.
func fetchAOMMetrics(ctx context.Context, aomClient *golangsdk.ServiceClient, hostID string) (*aomResponse, error) {
	queries, err := discoverAOMMetrics(ctx, aomClient, hostID)
	if err != nil {
		return nil, err
	}
	if len(queries) == 0 {
		return &aomResponse{}, nil
	}

	window, err := internal.NewSliceWindow(queries, Config.AOMBatchSize)
	if err != nil {
		return nil, fmt.Errorf("slice window: %w", err)
	}

	// Build all batches upfront so we can fetch them in parallel.
	var batches [][]aomMetricQuery
	for window.HasNext() {
		batches = append(batches, window.Window())
		window.NextWindow()
	}

	type batchResult struct {
		metrics []aomMetricResult
		err     error
	}
	results := make([]batchResult, len(batches))

	var wg sync.WaitGroup
	sem := make(chan struct{}, Config.AOMConcurrency)
	for i, batch := range batches {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, batch []aomMetricQuery) {
			defer wg.Done()
			defer func() { <-sem }()

			body := aomRequest{
				Metrics:    batch,
				Period:     60,
				Timerange:  "-1.-1.5",
				Statistics: []string{"average"},
			}

			var resp aomResponse
			_, err := aomClient.Post(aomClient.ServiceURL("ams", "metricdata"), body, &resp, &golangsdk.RequestOpts{
				OkCodes: []int{200},
			})
			if err != nil {
				results[idx] = batchResult{err: fmt.Errorf("aom api: %w", err)}
				return
			}
			if err := checkAOMError(resp.ErrorCode, resp.ErrorMessage); err != nil {
				results[idx] = batchResult{err: err}
				return
			}
			results[idx] = batchResult{metrics: resp.Metrics}
		}(i, batch)
	}
	wg.Wait()

	var allResults []aomMetricResult
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		allResults = append(allResults, r.metrics...)
	}

	return &aomResponse{Metrics: allResults}, nil
}

// convertAOMResponseToFamilies converts an AOM response to Prometheus MetricFamily
// objects. The instanceID and instanceName are used for resource_id/resource_name
// labels so AOM metrics join naturally with CES metrics in Grafana.
func convertAOMResponseToFamilies(resp *aomResponse, instanceID, instanceName string) []*dto.MetricFamily {
	if resp == nil {
		return nil
	}

	familyMap := make(map[string]*dto.MetricFamily)

	for _, result := range resp.Metrics {
		if len(result.Datapoints) == 0 {
			continue
		}

		// Find the latest datapoint by timestamp (not array position),
		// matching the CES pattern in ces.go.
		latest := result.Datapoints[0]
		for _, dp := range result.Datapoints[1:] {
			if dp.Timestamp > latest.Timestamp {
				latest = dp
			}
		}

		// Extract the "average" statistic value. If not found, skip —
		// this should not happen since we request "average", but is
		// safer than using a sentinel value.
		found := false
		var value float64
		for _, s := range latest.Statistics {
			if s.Statistic == "average" {
				value = s.Value
				found = true
				break
			}
		}
		if !found {
			continue
		}

		// The AOM API returns -1.0 as a fill value when no data exists
		// for a metric (controlled by the fillValue query parameter,
		// which defaults to -1). This happens for hosts not monitored
		// by AOM, e.g. standalone ECS instances that aren't part of a
		// CCE cluster. We skip these to avoid exposing misleading values.
		if value == -1.0 {
			continue
		}

		promName := aomPromName(result.Metric.MetricName)

		labels := map[string]string{
			"resource_id":   instanceID,
			"resource_name": instanceName,
		}

		// Add non-host dimension labels (diskDevice, netDevice, mountPoint, etc.)
		// auto-converted from camelCase to snake_case.
		for _, dim := range result.Metric.Dimensions {
			if aomHostDimensions[dim.Name] {
				continue
			}
			labels[camelToSnake(dim.Name)] = dim.Value
		}

		fam, exists := familyMap[promName]
		if !exists {
			fam = NewGaugeMetricFamily(promName, nil)
			familyMap[promName] = fam
		}
		fam.Metric = append(fam.Metric, NewGaugeMetric(value, labels))
	}

	families := make([]*dto.MetricFamily, 0, len(familyMap))
	for _, fam := range familyMap {
		families = append(families, fam)
	}
	return families
}

// aomPromName converts an AOM metric name to a Prometheus metric name.
// camelCase is converted to snake_case with the "ecs_aom_" prefix:
// "cpuUsage" → "ecs_aom_cpu_usage".
func aomPromName(aomName string) string {
	return "ecs_aom_" + camelToSnake(aomName)
}

// camelToSnake converts a camelCase string to snake_case.
// Each uppercase letter gets a preceding underscore and is lowercased.
// Consecutive uppercase letters are each separated: "diskRWStatus" → "disk_r_w_status".
func camelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
