//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination archival_metadata_mock.go

package archiver

import (
	"fmt"
	"strings"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/dynamicconfig"
)

type (
	// ArchivalMetadata provides cluster level archival information
	ArchivalMetadata interface {
		GetHistoryConfig() ArchivalConfig
		GetVisibilityConfig() ArchivalConfig
	}

	// ArchivalConfig is an immutable representation of the archival configuration of the cluster
	// This config is determined at cluster startup time
	ArchivalConfig interface {
		ClusterConfiguredForArchival() bool
		GetClusterState() ArchivalState
		ReadEnabled() bool
		GetNamespaceDefaultState() enumspb.ArchivalState
		GetNamespaceDefaultURI() string
		StaticClusterState() ArchivalState
	}

	archivalMetadata struct {
		historyConfig    ArchivalConfig
		visibilityConfig ArchivalConfig
	}

	archivalConfig struct {
		staticClusterState    ArchivalState
		dynamicClusterState   dynamicconfig.StringPropertyFn
		enableRead            dynamicconfig.BoolPropertyFn
		namespaceDefaultState enumspb.ArchivalState
		namespaceDefaultURI   string
	}

	// ArchivalState represents the archival state of the cluster
	ArchivalState int
)

func (a *archivalConfig) StaticClusterState() ArchivalState {
	return a.staticClusterState
}

const (
	// ArchivalDisabled means this cluster is not configured to handle archival
	ArchivalDisabled ArchivalState = iota
	// ArchivalPaused means this cluster is configured to handle archival but is currently not archiving
	// This state is not yet implemented, as of now ArchivalPaused is treated the same way as ArchivalDisabled
	ArchivalPaused
	// ArchivalEnabled means this cluster is currently archiving
	ArchivalEnabled
)

// NewArchivalMetadata constructs a new ArchivalMetadata
func NewArchivalMetadata(
	dc *dynamicconfig.Collection,
	historyState string,
	historyReadEnabled bool,
	visibilityState string,
	visibilityReadEnabled bool,
	namespaceDefaults *config.ArchivalNamespaceDefaults,
) ArchivalMetadata {
	historyConfig := NewArchivalConfig(
		historyState,
		dynamicconfig.HistoryArchivalState.WithDefault(historyState).Get(dc),
		dynamicconfig.EnableReadFromHistoryArchival.WithDefault(historyReadEnabled).Get(dc),
		namespaceDefaults.History.State,
		namespaceDefaults.History.URI,
	)

	visibilityConfig := NewArchivalConfig(
		visibilityState,
		dynamicconfig.VisibilityArchivalState.WithDefault(visibilityState).Get(dc),
		dynamicconfig.EnableReadFromVisibilityArchival.WithDefault(visibilityReadEnabled).Get(dc),
		namespaceDefaults.Visibility.State,
		namespaceDefaults.Visibility.URI,
	)

	return &archivalMetadata{
		historyConfig:    historyConfig,
		visibilityConfig: visibilityConfig,
	}
}

func (metadata *archivalMetadata) GetHistoryConfig() ArchivalConfig {
	return metadata.historyConfig
}

func (metadata *archivalMetadata) GetVisibilityConfig() ArchivalConfig {
	return metadata.visibilityConfig
}

// NewArchivalConfig constructs a new valid ArchivalConfig
func NewArchivalConfig(
	staticClusterStateStr string,
	dynamicClusterState dynamicconfig.StringPropertyFn,
	enableRead dynamicconfig.BoolPropertyFn,
	namespaceDefaultStateStr string,
	namespaceDefaultURI string,
) ArchivalConfig {
	staticClusterState, err := getClusterArchivalState(staticClusterStateStr)
	if err != nil {
		panic(err)
	}
	namespaceDefaultState, err := getNamespaceArchivalState(namespaceDefaultStateStr)
	if err != nil {
		panic(err)
	}

	return &archivalConfig{
		staticClusterState:    staticClusterState,
		dynamicClusterState:   dynamicClusterState,
		enableRead:            enableRead,
		namespaceDefaultState: namespaceDefaultState,
		namespaceDefaultURI:   namespaceDefaultURI,
	}
}

// NewDisabledArchvialConfig returns an ArchivalConfig where archival is disabled for both the cluster and the namespace
func NewDisabledArchvialConfig() ArchivalConfig {
	return &archivalConfig{
		staticClusterState:    ArchivalDisabled,
		dynamicClusterState:   nil,
		enableRead:            nil,
		namespaceDefaultState: enumspb.ARCHIVAL_STATE_DISABLED,
		namespaceDefaultURI:   "",
	}
}

// NewEnabledArchivalConfig returns an ArchivalConfig where archival is enabled for both the cluster and the namespace
func NewEnabledArchivalConfig() ArchivalConfig {
	return &archivalConfig{
		staticClusterState:    ArchivalEnabled,
		dynamicClusterState:   dynamicconfig.GetStringPropertyFn("enabled"),
		enableRead:            dynamicconfig.GetBoolPropertyFn(true),
		namespaceDefaultState: enumspb.ARCHIVAL_STATE_ENABLED,
		namespaceDefaultURI:   "some-uri",
	}
}

// ClusterConfiguredForArchival returns true if cluster is configured to handle archival, false otherwise
func (a *archivalConfig) ClusterConfiguredForArchival() bool {
	return a.GetClusterState() == ArchivalEnabled
}

func (a *archivalConfig) GetClusterState() ArchivalState {
	// Only check dynamic config when archival is enabled in static config.
	// If archival is disabled in static config, there will be no provider section in the static config
	// and the archiver provider can not create any archiver. Therefore, in that case,
	// even dynamic config says archival is enabled, we should ignore that.
	// Only when archival is enabled in static config, should we check if there's any difference between static config and dynamic config.
	if a.staticClusterState != ArchivalEnabled {
		return a.staticClusterState
	}

	dynamicStateStr := a.dynamicClusterState()
	dynamicState, err := getClusterArchivalState(dynamicStateStr)
	if err != nil {
		return ArchivalDisabled
	}
	return dynamicState
}

func (a *archivalConfig) ReadEnabled() bool {
	if !a.ClusterConfiguredForArchival() {
		return false
	}
	return a.enableRead()
}

func (a *archivalConfig) GetNamespaceDefaultState() enumspb.ArchivalState {
	return a.namespaceDefaultState
}

func (a *archivalConfig) GetNamespaceDefaultURI() string {
	return a.namespaceDefaultURI
}

func getClusterArchivalState(str string) (ArchivalState, error) {
	str = strings.TrimSpace(strings.ToLower(str))
	switch str {
	case "", config.ArchivalDisabled:
		return ArchivalDisabled, nil
	case config.ArchivalPaused:
		return ArchivalPaused, nil
	case config.ArchivalEnabled:
		return ArchivalEnabled, nil
	}
	return ArchivalDisabled, fmt.Errorf("invalid archival state of %v for cluster, valid states are: {\"\", \"disabled\", \"paused\", \"enabled\"}", str)
}

func getNamespaceArchivalState(str string) (enumspb.ArchivalState, error) {
	str = strings.TrimSpace(strings.ToLower(str))
	switch str {
	case "", config.ArchivalDisabled:
		return enumspb.ARCHIVAL_STATE_DISABLED, nil
	case config.ArchivalEnabled:
		return enumspb.ARCHIVAL_STATE_ENABLED, nil
	}
	return enumspb.ARCHIVAL_STATE_DISABLED, fmt.Errorf("invalid archival state of %v for namespace, valid states are: {\"\", \"disabled\", \"enabled\"}", str)
}
