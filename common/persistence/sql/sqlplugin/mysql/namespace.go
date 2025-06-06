package mysql

import (
	"context"
	"database/sql"
	"errors"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/server/common/persistence/sql/sqlplugin"
)

const (
	createNamespaceQuery = `INSERT INTO 
 namespaces (partition_id, id, name, is_global, data, data_encoding, notification_version)
 VALUES(?, ?, ?, ?, ?, ?, ?)`

	updateNamespaceQuery = `UPDATE namespaces 
 SET name = ?, data = ?, data_encoding = ?, is_global = ?, notification_version = ?
 WHERE partition_id=54321 AND id = ?`

	getNamespacePart = `SELECT id, name, is_global, data, data_encoding, notification_version FROM namespaces`

	getNamespaceByIDQuery   = getNamespacePart + ` WHERE partition_id=? AND id = ?`
	getNamespaceByNameQuery = getNamespacePart + ` WHERE partition_id=? AND name = ?`

	listNamespacesQuery      = getNamespacePart + ` WHERE partition_id=? ORDER BY id LIMIT ?`
	listNamespacesRangeQuery = getNamespacePart + ` WHERE partition_id=? AND id > ? ORDER BY id LIMIT ?`

	deleteNamespaceByIDQuery   = `DELETE FROM namespaces WHERE partition_id=? AND id = ?`
	deleteNamespaceByNameQuery = `DELETE FROM namespaces WHERE partition_id=? AND name = ?`

	getNamespaceMetadataQuery    = `SELECT notification_version FROM namespace_metadata WHERE partition_id = 54321`
	lockNamespaceMetadataQuery   = `SELECT notification_version FROM namespace_metadata WHERE partition_id = 54321 FOR UPDATE`
	updateNamespaceMetadataQuery = `UPDATE namespace_metadata SET notification_version = ? WHERE notification_version = ? AND partition_id = 54321`
)

const (
	partitionID = 54321
)

var errMissingArgs = errors.New("missing one or more args for API")

// InsertIntoNamespace inserts a single row into namespaces table
func (mdb *db) InsertIntoNamespace(
	ctx context.Context,
	row *sqlplugin.NamespaceRow,
) (sql.Result, error) {
	return mdb.ExecContext(ctx,
		createNamespaceQuery,
		partitionID,
		row.ID,
		row.Name,
		row.IsGlobal,
		row.Data,
		row.DataEncoding,
		row.NotificationVersion,
	)
}

// UpdateNamespace updates a single row in namespaces table
func (mdb *db) UpdateNamespace(
	ctx context.Context,
	row *sqlplugin.NamespaceRow,
) (sql.Result, error) {
	return mdb.ExecContext(ctx,
		updateNamespaceQuery,
		row.Name,
		row.Data,
		row.DataEncoding,
		row.IsGlobal,
		row.NotificationVersion,
		row.ID,
	)
}

// SelectFromNamespace reads one or more rows from namespaces table
func (mdb *db) SelectFromNamespace(
	ctx context.Context,
	filter sqlplugin.NamespaceFilter,
) ([]sqlplugin.NamespaceRow, error) {
	switch {
	case filter.ID != nil || filter.Name != nil:
		if filter.ID != nil && filter.Name != nil {
			return nil, serviceerror.NewInternal("only ID or name filter can be specified for selection")
		}
		return mdb.selectFromNamespace(ctx, filter)
	case filter.PageSize != nil && *filter.PageSize > 0:
		return mdb.selectAllFromNamespace(ctx, filter)
	default:
		return nil, errMissingArgs
	}
}

func (mdb *db) selectFromNamespace(
	ctx context.Context,
	filter sqlplugin.NamespaceFilter,
) ([]sqlplugin.NamespaceRow, error) {
	var err error
	var row sqlplugin.NamespaceRow
	switch {
	case filter.ID != nil:
		err = mdb.GetContext(ctx,
			&row,
			getNamespaceByIDQuery,
			partitionID,
			*filter.ID,
		)
	case filter.Name != nil:
		err = mdb.GetContext(ctx,
			&row,
			getNamespaceByNameQuery,
			partitionID,
			*filter.Name,
		)
	}
	if err != nil {
		return nil, err
	}
	return []sqlplugin.NamespaceRow{row}, nil
}

func (mdb *db) selectAllFromNamespace(
	ctx context.Context,
	filter sqlplugin.NamespaceFilter,
) ([]sqlplugin.NamespaceRow, error) {
	var err error
	var rows []sqlplugin.NamespaceRow
	switch {
	case filter.GreaterThanID != nil:
		err = mdb.SelectContext(ctx,
			&rows,
			listNamespacesRangeQuery,
			partitionID,
			*filter.GreaterThanID,
			*filter.PageSize,
		)
	default:
		err = mdb.SelectContext(ctx,
			&rows,
			listNamespacesQuery,
			partitionID,
			filter.PageSize,
		)
	}
	return rows, err
}

// DeleteFromNamespace deletes a single row in namespaces table
func (mdb *db) DeleteFromNamespace(
	ctx context.Context,
	filter sqlplugin.NamespaceFilter,
) (sql.Result, error) {
	var err error
	var result sql.Result
	switch {
	case filter.ID != nil:
		result, err = mdb.ExecContext(ctx,
			deleteNamespaceByIDQuery,
			partitionID,
			filter.ID,
		)
	default:
		result, err = mdb.ExecContext(ctx,
			deleteNamespaceByNameQuery,
			partitionID,
			filter.Name,
		)
	}
	return result, err
}

// LockNamespaceMetadata acquires a write lock on a single row in namespace_metadata table
func (mdb *db) LockNamespaceMetadata(
	ctx context.Context,
) (*sqlplugin.NamespaceMetadataRow, error) {
	var row sqlplugin.NamespaceMetadataRow
	err := mdb.GetContext(ctx,
		&row.NotificationVersion,
		lockNamespaceMetadataQuery,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// SelectFromNamespaceMetadata reads a single row in namespace_metadata table
func (mdb *db) SelectFromNamespaceMetadata(
	ctx context.Context,
) (*sqlplugin.NamespaceMetadataRow, error) {
	var row sqlplugin.NamespaceMetadataRow
	err := mdb.GetContext(ctx,
		&row.NotificationVersion,
		getNamespaceMetadataQuery,
	)
	return &row, err
}

// UpdateNamespaceMetadata updates a single row in namespace_metadata table
func (mdb *db) UpdateNamespaceMetadata(
	ctx context.Context,
	row *sqlplugin.NamespaceMetadataRow,
) (sql.Result, error) {
	return mdb.ExecContext(ctx,
		updateNamespaceMetadataQuery,
		row.NotificationVersion+1,
		row.NotificationVersion,
	)
}
