// Code generated by gowrap. DO NOT EDIT.
// template: gowrap_template
// gowrap: http://github.com/hexdigest/gowrap

package telemetry

//go:generate gowrap gen -p go.temporal.io/server/common/persistence -i ClusterMetadataStore -t gowrap_template -o cluster_metadata_store_gen.go -l ""

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	_sourcePersistence "go.temporal.io/server/common/persistence"
	"go.temporal.io/server/common/telemetry"
)

// telemetryClusterMetadataStore implements ClusterMetadataStore interface instrumented with OpenTelemetry.
type telemetryClusterMetadataStore struct {
	_sourcePersistence.ClusterMetadataStore
	tracer    trace.Tracer
	logger    log.Logger
	debugMode bool
}

// newTelemetryClusterMetadataStore returns telemetryClusterMetadataStore.
func newTelemetryClusterMetadataStore(
	base _sourcePersistence.ClusterMetadataStore,
	logger log.Logger,
	tracer trace.Tracer,
) telemetryClusterMetadataStore {
	return telemetryClusterMetadataStore{
		ClusterMetadataStore: base,
		tracer:               tracer,
		debugMode:            telemetry.DebugMode(),
	}
}

// DeleteClusterMetadata wraps ClusterMetadataStore.DeleteClusterMetadata.
func (d telemetryClusterMetadataStore) DeleteClusterMetadata(ctx context.Context, request *_sourcePersistence.InternalDeleteClusterMetadataRequest) (err error) {
	ctx, span := d.tracer.Start(
		ctx,
		"persistence.ClusterMetadataStore/DeleteClusterMetadata",
		trace.WithAttributes(
			attribute.Key("persistence.store").String("ClusterMetadataStore"),
			attribute.Key("persistence.method").String("DeleteClusterMetadata"),
		))
	defer span.End()

	if deadline, ok := ctx.Deadline(); ok {
		span.SetAttributes(attribute.String("deadline", deadline.Format(time.RFC3339Nano)))
		span.SetAttributes(attribute.String("timeout", time.Until(deadline).String()))
	}

	err = d.ClusterMetadataStore.DeleteClusterMetadata(ctx, request)
	if err != nil {
		span.RecordError(err)
	}

	if d.debugMode {

		requestPayload, err := json.MarshalIndent(request, "", "    ")
		if err != nil {
			d.logger.Error("failed to serialize *_sourcePersistence.InternalDeleteClusterMetadataRequest for OTEL span", tag.Error(err))
		} else {
			span.SetAttributes(attribute.Key("persistence.request.payload").String(string(requestPayload)))
		}

	}

	return
}

// GetClusterMembers wraps ClusterMetadataStore.GetClusterMembers.
func (d telemetryClusterMetadataStore) GetClusterMembers(ctx context.Context, request *_sourcePersistence.GetClusterMembersRequest) (gp1 *_sourcePersistence.GetClusterMembersResponse, err error) {
	ctx, span := d.tracer.Start(
		ctx,
		"persistence.ClusterMetadataStore/GetClusterMembers",
		trace.WithAttributes(
			attribute.Key("persistence.store").String("ClusterMetadataStore"),
			attribute.Key("persistence.method").String("GetClusterMembers"),
		))
	defer span.End()

	if deadline, ok := ctx.Deadline(); ok {
		span.SetAttributes(attribute.String("deadline", deadline.Format(time.RFC3339Nano)))
		span.SetAttributes(attribute.String("timeout", time.Until(deadline).String()))
	}

	gp1, err = d.ClusterMetadataStore.GetClusterMembers(ctx, request)
	if err != nil {
		span.RecordError(err)
	}

	if d.debugMode {

		requestPayload, err := json.MarshalIndent(request, "", "    ")
		if err != nil {
			d.logger.Error("failed to serialize *_sourcePersistence.GetClusterMembersRequest for OTEL span", tag.Error(err))
		} else {
			span.SetAttributes(attribute.Key("persistence.request.payload").String(string(requestPayload)))
		}

		responsePayload, err := json.MarshalIndent(gp1, "", "    ")
		if err != nil {
			d.logger.Error("failed to serialize *_sourcePersistence.GetClusterMembersResponse for OTEL span", tag.Error(err))
		} else {
			span.SetAttributes(attribute.Key("persistence.response.payload").String(string(responsePayload)))
		}

	}

	return
}

// GetClusterMetadata wraps ClusterMetadataStore.GetClusterMetadata.
func (d telemetryClusterMetadataStore) GetClusterMetadata(ctx context.Context, request *_sourcePersistence.InternalGetClusterMetadataRequest) (ip1 *_sourcePersistence.InternalGetClusterMetadataResponse, err error) {
	ctx, span := d.tracer.Start(
		ctx,
		"persistence.ClusterMetadataStore/GetClusterMetadata",
		trace.WithAttributes(
			attribute.Key("persistence.store").String("ClusterMetadataStore"),
			attribute.Key("persistence.method").String("GetClusterMetadata"),
		))
	defer span.End()

	if deadline, ok := ctx.Deadline(); ok {
		span.SetAttributes(attribute.String("deadline", deadline.Format(time.RFC3339Nano)))
		span.SetAttributes(attribute.String("timeout", time.Until(deadline).String()))
	}

	ip1, err = d.ClusterMetadataStore.GetClusterMetadata(ctx, request)
	if err != nil {
		span.RecordError(err)
	}

	if d.debugMode {

		requestPayload, err := json.MarshalIndent(request, "", "    ")
		if err != nil {
			d.logger.Error("failed to serialize *_sourcePersistence.InternalGetClusterMetadataRequest for OTEL span", tag.Error(err))
		} else {
			span.SetAttributes(attribute.Key("persistence.request.payload").String(string(requestPayload)))
		}

		responsePayload, err := json.MarshalIndent(ip1, "", "    ")
		if err != nil {
			d.logger.Error("failed to serialize *_sourcePersistence.InternalGetClusterMetadataResponse for OTEL span", tag.Error(err))
		} else {
			span.SetAttributes(attribute.Key("persistence.response.payload").String(string(responsePayload)))
		}

	}

	return
}

// ListClusterMetadata wraps ClusterMetadataStore.ListClusterMetadata.
func (d telemetryClusterMetadataStore) ListClusterMetadata(ctx context.Context, request *_sourcePersistence.InternalListClusterMetadataRequest) (ip1 *_sourcePersistence.InternalListClusterMetadataResponse, err error) {
	ctx, span := d.tracer.Start(
		ctx,
		"persistence.ClusterMetadataStore/ListClusterMetadata",
		trace.WithAttributes(
			attribute.Key("persistence.store").String("ClusterMetadataStore"),
			attribute.Key("persistence.method").String("ListClusterMetadata"),
		))
	defer span.End()

	if deadline, ok := ctx.Deadline(); ok {
		span.SetAttributes(attribute.String("deadline", deadline.Format(time.RFC3339Nano)))
		span.SetAttributes(attribute.String("timeout", time.Until(deadline).String()))
	}

	ip1, err = d.ClusterMetadataStore.ListClusterMetadata(ctx, request)
	if err != nil {
		span.RecordError(err)
	}

	if d.debugMode {

		requestPayload, err := json.MarshalIndent(request, "", "    ")
		if err != nil {
			d.logger.Error("failed to serialize *_sourcePersistence.InternalListClusterMetadataRequest for OTEL span", tag.Error(err))
		} else {
			span.SetAttributes(attribute.Key("persistence.request.payload").String(string(requestPayload)))
		}

		responsePayload, err := json.MarshalIndent(ip1, "", "    ")
		if err != nil {
			d.logger.Error("failed to serialize *_sourcePersistence.InternalListClusterMetadataResponse for OTEL span", tag.Error(err))
		} else {
			span.SetAttributes(attribute.Key("persistence.response.payload").String(string(responsePayload)))
		}

	}

	return
}

// PruneClusterMembership wraps ClusterMetadataStore.PruneClusterMembership.
func (d telemetryClusterMetadataStore) PruneClusterMembership(ctx context.Context, request *_sourcePersistence.PruneClusterMembershipRequest) (err error) {
	ctx, span := d.tracer.Start(
		ctx,
		"persistence.ClusterMetadataStore/PruneClusterMembership",
		trace.WithAttributes(
			attribute.Key("persistence.store").String("ClusterMetadataStore"),
			attribute.Key("persistence.method").String("PruneClusterMembership"),
		))
	defer span.End()

	if deadline, ok := ctx.Deadline(); ok {
		span.SetAttributes(attribute.String("deadline", deadline.Format(time.RFC3339Nano)))
		span.SetAttributes(attribute.String("timeout", time.Until(deadline).String()))
	}

	err = d.ClusterMetadataStore.PruneClusterMembership(ctx, request)
	if err != nil {
		span.RecordError(err)
	}

	if d.debugMode {

		requestPayload, err := json.MarshalIndent(request, "", "    ")
		if err != nil {
			d.logger.Error("failed to serialize *_sourcePersistence.PruneClusterMembershipRequest for OTEL span", tag.Error(err))
		} else {
			span.SetAttributes(attribute.Key("persistence.request.payload").String(string(requestPayload)))
		}

	}

	return
}

// SaveClusterMetadata wraps ClusterMetadataStore.SaveClusterMetadata.
func (d telemetryClusterMetadataStore) SaveClusterMetadata(ctx context.Context, request *_sourcePersistence.InternalSaveClusterMetadataRequest) (b1 bool, err error) {
	ctx, span := d.tracer.Start(
		ctx,
		"persistence.ClusterMetadataStore/SaveClusterMetadata",
		trace.WithAttributes(
			attribute.Key("persistence.store").String("ClusterMetadataStore"),
			attribute.Key("persistence.method").String("SaveClusterMetadata"),
		))
	defer span.End()

	if deadline, ok := ctx.Deadline(); ok {
		span.SetAttributes(attribute.String("deadline", deadline.Format(time.RFC3339Nano)))
		span.SetAttributes(attribute.String("timeout", time.Until(deadline).String()))
	}

	b1, err = d.ClusterMetadataStore.SaveClusterMetadata(ctx, request)
	if err != nil {
		span.RecordError(err)
	}

	if d.debugMode {

		requestPayload, err := json.MarshalIndent(request, "", "    ")
		if err != nil {
			d.logger.Error("failed to serialize *_sourcePersistence.InternalSaveClusterMetadataRequest for OTEL span", tag.Error(err))
		} else {
			span.SetAttributes(attribute.Key("persistence.request.payload").String(string(requestPayload)))
		}

		responsePayload, err := json.MarshalIndent(b1, "", "    ")
		if err != nil {
			d.logger.Error("failed to serialize bool for OTEL span", tag.Error(err))
		} else {
			span.SetAttributes(attribute.Key("persistence.response.payload").String(string(responsePayload)))
		}

	}

	return
}

// UpsertClusterMembership wraps ClusterMetadataStore.UpsertClusterMembership.
func (d telemetryClusterMetadataStore) UpsertClusterMembership(ctx context.Context, request *_sourcePersistence.UpsertClusterMembershipRequest) (err error) {
	ctx, span := d.tracer.Start(
		ctx,
		"persistence.ClusterMetadataStore/UpsertClusterMembership",
		trace.WithAttributes(
			attribute.Key("persistence.store").String("ClusterMetadataStore"),
			attribute.Key("persistence.method").String("UpsertClusterMembership"),
		))
	defer span.End()

	if deadline, ok := ctx.Deadline(); ok {
		span.SetAttributes(attribute.String("deadline", deadline.Format(time.RFC3339Nano)))
		span.SetAttributes(attribute.String("timeout", time.Until(deadline).String()))
	}

	err = d.ClusterMetadataStore.UpsertClusterMembership(ctx, request)
	if err != nil {
		span.RecordError(err)
	}

	if d.debugMode {

		requestPayload, err := json.MarshalIndent(request, "", "    ")
		if err != nil {
			d.logger.Error("failed to serialize *_sourcePersistence.UpsertClusterMembershipRequest for OTEL span", tag.Error(err))
		} else {
			span.SetAttributes(attribute.Key("persistence.request.payload").String(string(requestPayload)))
		}

	}

	return
}
