package interceptor

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/api/historyservice/v1"
	"go.temporal.io/server/api/matchingservice/v1"
	"go.temporal.io/server/common/namespace"
	"go.uber.org/mock/gomock"
)

type (
	namespaceSuite struct {
		suite.Suite
		*require.Assertions
	}
)

var (
	frontendAPIExcluded = map[string]struct{}{
		"GetClusterInfo":      {},
		"GetSystemInfo":       {},
		"GetSearchAttributes": {},
		"ListNamespaces":      {},
	}

	matchingAPIExcluded = map[string]struct{}{
		"ListTaskQueuePartitions": {},
		// Nexus endpoint APIs operate on a cluster scope, not a namespace scope.
		"CreateNexusEndpoint": {},
		"UpdateNexusEndpoint": {},
		"ListNexusEndpoints":  {},
		"DeleteNexusEndpoint": {},
	}

	historyAPIExcluded = map[string]struct{}{
		"CloseShard":                {},
		"GetShard":                  {},
		"GetDLQMessages":            {},
		"GetDLQReplicationMessages": {},
		"GetReplicationMessages":    {},
		"MergeDLQMessages":          {},
		"PurgeDLQMessages":          {},
		"RemoveTask":                {},
		"SyncShardStatus":           {},
		"GetReplicationStatus":      {},
		"GetDLQTasks":               {},
		"DeleteDLQTasks":            {},
		"AddTasks":                  {},
		"ListQueues":                {},
		"ListTasks":                 {},
		"CompleteNexusOperation":    {}, // NamespaceId is in the completion token for this request.
		"DeepHealthCheck":           {},
	}
)

func TestNamespaceSuite(t *testing.T) {
	s := new(namespaceSuite)
	suite.Run(t, s)
}

func (s *namespaceSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *namespaceSuite) TearDownTest() {

}

func (s *namespaceSuite) TestFrontendAPIMetrics() {
	namespaceNameGetter := reflect.TypeOf((*NamespaceNameGetter)(nil)).Elem()

	var service workflowservice.WorkflowServiceServer
	t := reflect.TypeOf(&service).Elem()
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methodName := method.Name
		methodType := method.Type

		// 0th parameter is context.Context
		// 1th parameter is the request
		if _, ok := frontendAPIExcluded[methodName]; ok {
			continue
		}
		if methodType.NumIn() < 2 {
			continue
		}
		request := methodType.In(1)
		if !request.Implements(namespaceNameGetter) {
			s.Fail(fmt.Sprintf("API: %v not implementing NamespaceNameGetter", methodName))
		}
	}
}

func (s *namespaceSuite) TestMatchingAPIMetrics() {
	namespaceIDGetter := reflect.TypeOf((*NamespaceIDGetter)(nil)).Elem()

	var service matchingservice.MatchingServiceServer
	t := reflect.TypeOf(&service).Elem()
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methodName := method.Name
		methodType := method.Type

		// 0th parameter is context.Context
		// 1th parameter is the request
		if _, ok := matchingAPIExcluded[methodName]; ok {
			continue
		}
		if methodType.NumIn() < 2 {
			continue
		}
		request := methodType.In(1)
		if !request.Implements(namespaceIDGetter) {
			s.Fail(fmt.Sprintf("API: %v not implementing NamespaceIDGetter", methodName))
		}
	}
}

func (s *namespaceSuite) TestHistoryAPIMetrics() {
	namespaceIDGetter := reflect.TypeOf((*NamespaceIDGetter)(nil)).Elem()

	var service historyservice.HistoryServiceServer
	t := reflect.TypeOf(&service).Elem()
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methodName := method.Name
		methodType := method.Type

		// 0th parameter is context.Context
		// 1th parameter is the request
		if _, ok := historyAPIExcluded[methodName]; ok {
			continue
		}
		if methodType.NumIn() < 2 {
			continue
		}
		request := methodType.In(1)
		if !request.Implements(namespaceIDGetter) {
			s.Fail(fmt.Sprintf("API: %v not implementing NamespaceIDGetter", methodName))
		}
	}
}

func (s *namespaceSuite) TestGetNamespace() {
	register := namespace.NewMockRegistry(gomock.NewController(s.T()))
	register.EXPECT().GetNamespace(namespace.Name("exist")).Return(nil, nil)
	register.EXPECT().GetNamespace(namespace.Name("nonexist")).Return(nil, errors.New("not found"))
	register.EXPECT().GetNamespaceName(namespace.ID("exist")).Return(namespace.Name("exist"), nil)
	register.EXPECT().GetNamespaceName(namespace.ID("nonexist")).Return(namespace.EmptyName, errors.New("not found"))
	testCases := []struct {
		method        interface{}
		namespaceName namespace.Name
	}{
		{
			&workflowservice.DescribeNamespaceRequest{Namespace: "exist"},
			namespace.Name("exist"),
		},
		{
			&workflowservice.DescribeNamespaceRequest{Namespace: "nonexist"},
			namespace.EmptyName,
		},
		{
			&historyservice.DescribeMutableStateRequest{NamespaceId: "exist"},
			namespace.Name("exist"),
		},
		{
			&historyservice.DescribeMutableStateRequest{NamespaceId: "nonexist"},
			namespace.EmptyName,
		},
	}

	for _, testCase := range testCases {
		extractedNamespace := MustGetNamespaceName(register, testCase.method)
		s.Equal(testCase.namespaceName, extractedNamespace)
	}
}
