//go:generate mockgen -package $GOPACKAGE -source query_parser.go -destination query_parser_mock.go -mock_names Interface=MockQueryParser

package s3store

import (
	"errors"
	"fmt"
	"time"

	"github.com/temporalio/sqlparser"
	"go.temporal.io/server/common/sqlquery"
	"go.temporal.io/server/common/util"
)

type (
	// QueryParser parses a limited SQL where clause into a struct
	QueryParser interface {
		Parse(query string) (*parsedQuery, error)
	}

	queryParser struct{}

	parsedQuery struct {
		workflowTypeName *string
		workflowID       *string
		startTime        *time.Time
		closeTime        *time.Time
		searchPrecision  *string
	}
)

// All allowed fields for filtering
const (
	WorkflowTypeName = "WorkflowTypeName"
	WorkflowID       = "WorkflowId"
	StartTime        = "StartTime"
	CloseTime        = "CloseTime"
	SearchPrecision  = "SearchPrecision"
)

// Precision specific values
const (
	PrecisionDay    = "Day"
	PrecisionHour   = "Hour"
	PrecisionMinute = "Minute"
	PrecisionSecond = "Second"
)

// NewQueryParser creates a new query parser for filestore
func NewQueryParser() QueryParser {
	return &queryParser{}
}

func (p *queryParser) Parse(query string) (*parsedQuery, error) {
	stmt, err := sqlparser.Parse(fmt.Sprintf(sqlquery.QueryTemplate, query))
	if err != nil {
		return nil, err
	}
	whereExpr := stmt.(*sqlparser.Select).Where.Expr
	parsedQuery := &parsedQuery{}
	if err := p.convertWhereExpr(whereExpr, parsedQuery); err != nil {
		return nil, err
	}
	if parsedQuery.workflowID == nil && parsedQuery.workflowTypeName == nil {
		return nil, errors.New("WorkflowId or WorkflowTypeName is required in query")
	}
	if parsedQuery.workflowID != nil && parsedQuery.workflowTypeName != nil {
		return nil, errors.New("only one of WorkflowId or WorkflowTypeName can be specified in a query")
	}
	if parsedQuery.closeTime != nil && parsedQuery.startTime != nil {
		return nil, errors.New("only one of StartTime or CloseTime can be specified in a query")
	}
	if (parsedQuery.closeTime != nil || parsedQuery.startTime != nil) && parsedQuery.searchPrecision == nil {
		return nil, errors.New("SearchPrecision is required when searching for a StartTime or CloseTime")
	}

	if parsedQuery.closeTime == nil && parsedQuery.startTime == nil && parsedQuery.searchPrecision != nil {
		return nil, errors.New("SearchPrecision requires a StartTime or CloseTime")
	}
	return parsedQuery, nil
}

func (p *queryParser) convertWhereExpr(expr sqlparser.Expr, parsedQuery *parsedQuery) error {
	if expr == nil {
		return errors.New("where expression is nil")
	}

	switch expr := expr.(type) {
	case *sqlparser.ComparisonExpr:
		return p.convertComparisonExpr(expr, parsedQuery)
	case *sqlparser.AndExpr:
		return p.convertAndExpr(expr, parsedQuery)
	case *sqlparser.ParenExpr:
		return p.convertParenExpr(expr, parsedQuery)
	default:
		return errors.New("only comparison and \"and\" expression is supported")
	}
}

func (p *queryParser) convertParenExpr(parenExpr *sqlparser.ParenExpr, parsedQuery *parsedQuery) error {
	return p.convertWhereExpr(parenExpr.Expr, parsedQuery)
}

func (p *queryParser) convertAndExpr(andExpr *sqlparser.AndExpr, parsedQuery *parsedQuery) error {
	if err := p.convertWhereExpr(andExpr.Left, parsedQuery); err != nil {
		return err
	}
	return p.convertWhereExpr(andExpr.Right, parsedQuery)
}

func (p *queryParser) convertComparisonExpr(compExpr *sqlparser.ComparisonExpr, parsedQuery *parsedQuery) error {
	colName, ok := compExpr.Left.(*sqlparser.ColName)
	if !ok {
		return fmt.Errorf("invalid filter name: %s", sqlparser.String(compExpr.Left))
	}
	colNameStr := sqlparser.String(colName)
	op := compExpr.Operator
	valExpr, ok := compExpr.Right.(*sqlparser.SQLVal)
	if !ok {
		return fmt.Errorf("invalid value: %s", sqlparser.String(compExpr.Right))
	}
	valStr := sqlparser.String(valExpr)

	switch colNameStr {
	case WorkflowTypeName:
		val, err := sqlquery.ExtractStringValue(valStr)
		if err != nil {
			return err
		}
		if op != "=" {
			return fmt.Errorf("only operation = is support for %s", WorkflowTypeName)
		}
		if parsedQuery.workflowTypeName != nil {
			return fmt.Errorf("can not query %s multiple times", WorkflowTypeName)
		}
		parsedQuery.workflowTypeName = util.Ptr(val)
	case WorkflowID:
		val, err := sqlquery.ExtractStringValue(valStr)
		if err != nil {
			return err
		}
		if op != "=" {
			return fmt.Errorf("only operation = is support for %s", WorkflowID)
		}
		if parsedQuery.workflowID != nil {
			return fmt.Errorf("can not query %s multiple times", WorkflowID)
		}
		parsedQuery.workflowID = util.Ptr(val)
	case CloseTime:
		timestamp, err := sqlquery.ConvertToTime(valStr)
		if err != nil {
			return err
		}
		if op != "=" {
			return fmt.Errorf("only operation = is support for %s", CloseTime)
		}
		parsedQuery.closeTime = &timestamp
	case StartTime:
		timestamp, err := sqlquery.ConvertToTime(valStr)
		if err != nil {
			return err
		}
		if op != "=" {
			return fmt.Errorf("only operation = is support for %s", CloseTime)
		}
		parsedQuery.startTime = &timestamp
	case SearchPrecision:
		val, err := sqlquery.ExtractStringValue(valStr)
		if err != nil {
			return err
		}
		if op != "=" {
			return fmt.Errorf("only operation = is support for %s", SearchPrecision)
		}
		if parsedQuery.searchPrecision != nil && *parsedQuery.searchPrecision != val {
			return fmt.Errorf("only one expression is allowed for %s", SearchPrecision)
		}
		switch val {
		case PrecisionDay:
		case PrecisionHour:
		case PrecisionMinute:
		case PrecisionSecond:
		default:
			return fmt.Errorf("invalid value for %s: %s", SearchPrecision, val)
		}
		parsedQuery.searchPrecision = util.Ptr(val)

	default:
		return fmt.Errorf("unknown filter name: %s", colNameStr)
	}

	return nil
}
