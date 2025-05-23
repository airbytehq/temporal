package filestore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/server/common/util"
)

type queryParserSuite struct {
	*require.Assertions
	suite.Suite

	parser QueryParser
}

func TestQueryParserSuite(t *testing.T) {
	suite.Run(t, new(queryParserSuite))
}

func (s *queryParserSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.parser = NewQueryParser()
}

func (s *queryParserSuite) TestParseWorkflowID_RunID_WorkflowType() {
	testCases := []struct {
		query       string
		expectErr   bool
		parsedQuery *parsedQuery
	}{
		{
			query:     "WorkflowId = \"random workflowID\"",
			expectErr: false,
			parsedQuery: &parsedQuery{
				workflowID: util.Ptr("random workflowID"),
			},
		},
		{
			query:     "WorkflowId = \"random workflowID\" and WorkflowId = \"random workflowID\"",
			expectErr: false,
			parsedQuery: &parsedQuery{
				workflowID: util.Ptr("random workflowID"),
			},
		},
		{
			query:     "RunId = \"random runID\"",
			expectErr: false,
			parsedQuery: &parsedQuery{
				runID: util.Ptr("random runID"),
			},
		},
		{
			query:     "WorkflowType = \"random typeName\"",
			expectErr: false,
			parsedQuery: &parsedQuery{
				workflowTypeName: util.Ptr("random typeName"),
			},
		},
		{
			query:     "WorkflowId = 'random workflowID'",
			expectErr: false,
			parsedQuery: &parsedQuery{
				workflowID: util.Ptr("random workflowID"),
			},
		},
		{
			query:     "WorkflowType = 'random typeName' and WorkflowType = \"another typeName\"",
			expectErr: false,
			parsedQuery: &parsedQuery{
				emptyResult: true,
			},
		},
		{
			query:     "WorkflowType = 'random typeName' and (WorkflowId = \"random workflowID\" and RunId='random runID')",
			expectErr: false,
			parsedQuery: &parsedQuery{
				workflowID:       util.Ptr("random workflowID"),
				runID:            util.Ptr("random runID"),
				workflowTypeName: util.Ptr("random typeName"),
			},
		},
		{
			query:     "runId = random workflowID",
			expectErr: true,
		},
		{
			query:     "WorkflowId = \"random workflowID\" or WorkflowId = \"another workflowID\"",
			expectErr: true,
		},
		{
			query:     "WorkflowId = \"random workflowID\" or runId = \"random runID\"",
			expectErr: true,
		},
		{
			query:     "workflowid = \"random workflowID\"",
			expectErr: true,
		},
		{
			query:     "runId > \"random workflowID\"",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		parsedQuery, err := s.parser.Parse(tc.query)
		if tc.expectErr {
			s.Error(err)
			continue
		}
		s.NoError(err)
		s.Equal(tc.parsedQuery.emptyResult, parsedQuery.emptyResult)
		if !tc.parsedQuery.emptyResult {
			s.Equal(tc.parsedQuery.workflowID, parsedQuery.workflowID)
			s.Equal(tc.parsedQuery.runID, parsedQuery.runID)
			s.Equal(tc.parsedQuery.workflowTypeName, parsedQuery.workflowTypeName)
		}
	}
}

func (s *queryParserSuite) TestParseCloseStatus() {
	testCases := []struct {
		query       string
		expectErr   bool
		parsedQuery *parsedQuery
	}{
		{
			query:     "ExecutionStatus = \"Completed\"",
			expectErr: false,
			parsedQuery: &parsedQuery{
				status: toWorkflowExecutionStatusPtr(enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED),
			},
		},
		{
			query:     "ExecutionStatus = \"failed\"",
			expectErr: false,
			parsedQuery: &parsedQuery{
				status: toWorkflowExecutionStatusPtr(enumspb.WORKFLOW_EXECUTION_STATUS_FAILED),
			},
		},
		{
			query:     "ExecutionStatus = \"canceled\"",
			expectErr: false,
			parsedQuery: &parsedQuery{
				status: toWorkflowExecutionStatusPtr(enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED),
			},
		},
		{
			query:     "ExecutionStatus = \"terminated\"",
			expectErr: false,
			parsedQuery: &parsedQuery{
				status: toWorkflowExecutionStatusPtr(enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED),
			},
		},
		{
			query:     "ExecutionStatus = 'continuedasnew'",
			expectErr: false,
			parsedQuery: &parsedQuery{
				status: toWorkflowExecutionStatusPtr(enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW),
			},
		},
		{
			query:     "ExecutionStatus = 'TIMED_OUT'",
			expectErr: false,
			parsedQuery: &parsedQuery{
				status: toWorkflowExecutionStatusPtr(enumspb.WORKFLOW_EXECUTION_STATUS_TIMED_OUT),
			},
		},
		{
			query:     "ExecutionStatus = 'Failed' and ExecutionStatus = \"Failed\"",
			expectErr: false,
			parsedQuery: &parsedQuery{
				status: toWorkflowExecutionStatusPtr(enumspb.WORKFLOW_EXECUTION_STATUS_FAILED),
			},
		},
		{
			query:     "(ExecutionStatus = 'Timedout' and ExecutionStatus = \"canceled\")",
			expectErr: false,
			parsedQuery: &parsedQuery{
				emptyResult: true,
			},
		},
		{
			query:     "status = \"Failed\"",
			expectErr: true,
		},
		{
			query:     "ExecutionStatus = \"Failed\" or ExecutionStatus = \"Failed\"",
			expectErr: true,
		},
		{
			query:     "ExecutionStatus = \"unknown\"",
			expectErr: true,
		},
		{
			query:     "ExecutionStatus > \"Failed\"",
			expectErr: true,
		},
		{
			query:     "ExecutionStatus = 3",
			expectErr: false,
			parsedQuery: &parsedQuery{
				status: toWorkflowExecutionStatusPtr(enumspb.WORKFLOW_EXECUTION_STATUS_FAILED),
			},
		},
		{
			query:     "CloseStatus = 10",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		parsedQuery, err := s.parser.Parse(tc.query)
		if tc.expectErr {
			s.Error(err)
			continue
		}
		s.NoError(err)
		s.Equal(tc.parsedQuery.emptyResult, parsedQuery.emptyResult)
		if !tc.parsedQuery.emptyResult {
			s.EqualValues(tc.parsedQuery.status, parsedQuery.status)
		}
	}
}

func (s *queryParserSuite) TestParseCloseTime() {
	testCases := []struct {
		query       string
		expectErr   bool
		parsedQuery *parsedQuery
	}{
		{
			query:     "CloseTime <= 1000",
			expectErr: false,
			parsedQuery: &parsedQuery{
				earliestCloseTime: time.Time{},
				latestCloseTime:   time.Unix(0, 1000),
			},
		},
		{
			query:     "CloseTime < 2000 and CloseTime <= 1000 and CloseTime > 300",
			expectErr: false,
			parsedQuery: &parsedQuery{
				earliestCloseTime: time.Unix(0, 301),
				latestCloseTime:   time.Unix(0, 1000),
			},
		},
		{
			query:     "CloseTime = 2000 and (CloseTime > 1000 and CloseTime <= 9999)",
			expectErr: false,
			parsedQuery: &parsedQuery{
				earliestCloseTime: time.Unix(0, 2000),
				latestCloseTime:   time.Unix(0, 2000),
			},
		},
		{
			query:     "CloseTime <= \"2019-01-01T11:11:11Z\" and CloseTime >= 1000000",
			expectErr: false,
			parsedQuery: &parsedQuery{
				earliestCloseTime: time.Unix(0, 1000000),
				latestCloseTime:   time.Date(2019, 01, 01, 11, 11, 11, 0, time.UTC),
			},
		},
		{
			query:     "closeTime = 2000",
			expectErr: true,
		},
		{
			query:     "CloseTime > \"2019-01-01 00:00:00\"",
			expectErr: true,
		},
		{
			query:     "ExecutionStatus > 2000 or ExecutionStatus < 1000",
			expectErr: true,
		},
	}

	for i, tc := range testCases {
		parsedQuery, err := s.parser.Parse(tc.query)
		if tc.expectErr {
			s.Error(err)
			continue
		}
		s.NoError(err, "case %d", i)
		s.Equal(tc.parsedQuery.emptyResult, parsedQuery.emptyResult, "case %d", i)
		if !tc.parsedQuery.emptyResult {
			s.True(tc.parsedQuery.earliestCloseTime.Equal(parsedQuery.earliestCloseTime), "case %d", i)
			s.True(tc.parsedQuery.latestCloseTime.Equal(parsedQuery.latestCloseTime), "case %d", i)
		}
	}
}

func (s *queryParserSuite) TestParse() {
	testCases := []struct {
		query       string
		expectErr   bool
		parsedQuery *parsedQuery
	}{
		{
			query:     "CloseTime <= \"2019-01-01T11:11:11Z\" and WorkflowId = 'random workflowID'",
			expectErr: false,
			parsedQuery: &parsedQuery{
				earliestCloseTime: time.Time{},
				latestCloseTime:   time.Date(2019, 01, 01, 11, 11, 11, 0, time.UTC),
				workflowID:        util.Ptr("random workflowID"),
			},
		},
		{
			query:     "CloseTime > 1999 and CloseTime < 10000 and RunId = 'random runID' and ExecutionStatus = 'Failed'",
			expectErr: false,
			parsedQuery: &parsedQuery{
				earliestCloseTime: time.Unix(0, 2000).UTC(),
				latestCloseTime:   time.Unix(0, 9999).UTC(),
				runID:             util.Ptr("random runID"),
				status:            toWorkflowExecutionStatusPtr(enumspb.WORKFLOW_EXECUTION_STATUS_FAILED),
			},
		},
		{
			query:     "CloseTime > 2001 and CloseTime < 10000 and (RunId = 'random runID') and ExecutionStatus = 'Failed' and (RunId = 'another ID')",
			expectErr: false,
			parsedQuery: &parsedQuery{
				emptyResult: true,
			},
		},
	}

	for i, tc := range testCases {
		parsedQuery, err := s.parser.Parse(tc.query)
		if tc.expectErr {
			s.Error(err)
			continue
		}
		s.NoError(err, "case %d", i)
		s.Equal(tc.parsedQuery.emptyResult, parsedQuery.emptyResult, "case %d", i)
		if !tc.parsedQuery.emptyResult {
			s.Equal(tc.parsedQuery, parsedQuery, "case %d", i)
		}
	}
}
