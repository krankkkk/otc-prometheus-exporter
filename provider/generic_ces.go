package provider

import (
	"context"
	"errors"
	"fmt"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// isCESClientError returns true if the error is a 4xx HTTP response from CES.
func isCESClientError(err error) bool {
	var err400 golangsdk.ErrDefault400
	var err404 golangsdk.ErrDefault404
	return errors.As(err, &err400) || errors.As(err, &err404)
}

// ErrNamespaceNotFound indicates CES does not recognise the requested namespace.
type ErrNamespaceNotFound struct {
	Namespace string
	Err       error
}

func (e *ErrNamespaceNotFound) Error() string {
	return fmt.Sprintf("namespace %q not found: %v", e.Namespace, e.Err)
}

func (e *ErrNamespaceNotFound) Unwrap() error { return e.Err }

// GenericCESProvider handles any CES namespace that does not have
// a dedicated registered provider. It calls CollectCESMetrics directly.
// If CES rejects the namespace with a 4xx error, the error is wrapped
// in ErrNamespaceNotFound so the handler can return 404 instead of 500.
type GenericCESProvider struct{ namespace string }

func (p *GenericCESProvider) Namespace() string { return p.namespace }

func (p *GenericCESProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	if client == nil {
		return nil, nil
	}
	families, err := CollectCESMetrics(ctx, client, p.namespace)
	if err != nil && isCESClientError(err) {
		return nil, &ErrNamespaceNotFound{Namespace: p.namespace, Err: err}
	}
	return families, err
}
