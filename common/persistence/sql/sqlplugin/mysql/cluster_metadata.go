package mysql

import (
	"context"
	"database/sql"
	"strings"

	p "go.temporal.io/server/common/persistence"
	"go.temporal.io/server/common/persistence/sql/sqlplugin"
)

const constMetadataPartition = 0
const constMembershipPartition = 0
const (
	// ****** CLUSTER_METADATA_INFO TABLE ******
	insertClusterMetadataQry = `INSERT INTO cluster_metadata_info (metadata_partition, cluster_name, data, data_encoding, version) VALUES(?, ?, ?, ?, ?)`

	updateClusterMetadataQry = `UPDATE cluster_metadata_info SET data = ?, data_encoding = ?, version = ? WHERE metadata_partition = ? AND cluster_name = ?`

	getClusterMetadataBase         = `SELECT data, data_encoding, version FROM cluster_metadata_info `
	listClusterMetadataQry         = getClusterMetadataBase + `WHERE metadata_partition = ? ORDER BY cluster_name LIMIT ?`
	listClusterMetadataRangeQry    = getClusterMetadataBase + `WHERE metadata_partition = ? AND cluster_name > ? ORDER BY cluster_name LIMIT ?`
	getClusterMetadataQry          = getClusterMetadataBase + `WHERE metadata_partition = ? AND cluster_name = ?`
	writeLockGetClusterMetadataQry = getClusterMetadataQry + ` FOR UPDATE`

	deleteClusterMetadataQry = `DELETE FROM cluster_metadata_info WHERE metadata_partition = ? AND cluster_name = ?`

	// ****** CLUSTER_MEMBERSHIP TABLE ******
	templateUpsertActiveClusterMembership = `INSERT INTO
cluster_membership (membership_partition, host_id, rpc_address, rpc_port, role, session_start, last_heartbeat, record_expiry)
VALUES(?, ?, ?, ?, ?, ?, ?, ?) 
ON DUPLICATE KEY UPDATE 
session_start=VALUES(session_start), last_heartbeat=VALUES(last_heartbeat), record_expiry=VALUES(record_expiry)`

	templatePruneStaleClusterMembership = `DELETE FROM
cluster_membership 
WHERE membership_partition = ? AND record_expiry < ?`

	templateGetClusterMembership = `SELECT host_id, rpc_address, rpc_port, role, session_start, last_heartbeat, record_expiry FROM
cluster_membership WHERE membership_partition = ?`

	// ClusterMembership WHERE Suffixes
	templateWithRoleSuffix           = ` AND role = ?`
	templateWithHeartbeatSinceSuffix = ` AND last_heartbeat > ?`
	templateWithRecordExpirySuffix   = ` AND record_expiry > ?`
	templateWithRPCAddressSuffix     = ` AND rpc_address = ?`
	templateWithHostIDSuffix         = ` AND host_id = ?`
	templateWithHostIDGreaterSuffix  = ` AND host_id > ?`
	templateWithSessionStartSuffix   = ` AND session_start >= ?`

	// Generic SELECT Suffixes
	templateWithLimitSuffix               = ` LIMIT ?`
	templateWithOrderBySessionStartSuffix = ` ORDER BY membership_partition ASC, host_id ASC`
)

func (mdb *db) SaveClusterMetadata(
	ctx context.Context,
	row *sqlplugin.ClusterMetadataRow,
) (sql.Result, error) {
	if row.Version == 0 {
		return mdb.ExecContext(ctx,
			insertClusterMetadataQry,
			constMetadataPartition,
			row.ClusterName,
			row.Data,
			row.DataEncoding,
			1,
		)
	}
	return mdb.ExecContext(ctx,
		updateClusterMetadataQry,
		row.Data,
		row.DataEncoding,
		row.Version+1,
		constMetadataPartition,
		row.ClusterName,
	)
}

func (mdb *db) ListClusterMetadata(
	ctx context.Context,
	filter *sqlplugin.ClusterMetadataFilter,
) ([]sqlplugin.ClusterMetadataRow, error) {
	var err error
	var rows []sqlplugin.ClusterMetadataRow
	switch {
	case len(filter.ClusterName) != 0:
		err = mdb.SelectContext(ctx,
			&rows,
			listClusterMetadataRangeQry,
			constMetadataPartition,
			filter.ClusterName,
			filter.PageSize,
		)
	default:
		err = mdb.SelectContext(ctx,
			&rows,
			listClusterMetadataQry,
			constMetadataPartition,
			filter.PageSize,
		)
	}
	return rows, err
}

func (mdb *db) GetClusterMetadata(
	ctx context.Context,
	filter *sqlplugin.ClusterMetadataFilter,
) (*sqlplugin.ClusterMetadataRow, error) {
	var row sqlplugin.ClusterMetadataRow
	err := mdb.GetContext(ctx,
		&row,
		getClusterMetadataQry,
		constMetadataPartition,
		filter.ClusterName,
	)
	if err != nil {
		return nil, err
	}
	return &row, err
}

func (mdb *db) DeleteClusterMetadata(
	ctx context.Context,
	filter *sqlplugin.ClusterMetadataFilter,
) (sql.Result, error) {

	return mdb.ExecContext(ctx,
		deleteClusterMetadataQry,
		constMetadataPartition,
		filter.ClusterName,
	)
}

func (mdb *db) WriteLockGetClusterMetadata(
	ctx context.Context,
	filter *sqlplugin.ClusterMetadataFilter,
) (*sqlplugin.ClusterMetadataRow, error) {
	var row sqlplugin.ClusterMetadataRow
	err := mdb.GetContext(ctx,
		&row,
		writeLockGetClusterMetadataQry,
		constMetadataPartition,
		filter.ClusterName,
	)
	if err != nil {
		return nil, err
	}
	return &row, err
}

func (mdb *db) UpsertClusterMembership(
	ctx context.Context,
	row *sqlplugin.ClusterMembershipRow,
) (sql.Result, error) {
	return mdb.ExecContext(ctx,
		templateUpsertActiveClusterMembership,
		constMembershipPartition,
		row.HostID,
		row.RPCAddress,
		row.RPCPort,
		row.Role,
		mdb.converter.ToMySQLDateTime(row.SessionStart),
		mdb.converter.ToMySQLDateTime(row.LastHeartbeat),
		mdb.converter.ToMySQLDateTime(row.RecordExpiry))
}

func (mdb *db) GetClusterMembers(
	ctx context.Context,
	filter *sqlplugin.ClusterMembershipFilter,
) ([]sqlplugin.ClusterMembershipRow, error) {
	var queryString strings.Builder
	var operands []interface{}
	queryString.WriteString(templateGetClusterMembership)
	operands = append(operands, constMembershipPartition)

	if filter.HostIDEquals != nil {
		queryString.WriteString(templateWithHostIDSuffix)
		operands = append(operands, filter.HostIDEquals)
	}

	if filter.RPCAddressEquals != "" {
		queryString.WriteString(templateWithRPCAddressSuffix)
		operands = append(operands, filter.RPCAddressEquals)
	}

	if filter.RoleEquals != p.All {
		queryString.WriteString(templateWithRoleSuffix)
		operands = append(operands, filter.RoleEquals)
	}

	if !filter.LastHeartbeatAfter.IsZero() {
		queryString.WriteString(templateWithHeartbeatSinceSuffix)
		operands = append(operands, filter.LastHeartbeatAfter)
	}

	if !filter.RecordExpiryAfter.IsZero() {
		queryString.WriteString(templateWithRecordExpirySuffix)
		operands = append(operands, filter.RecordExpiryAfter)
	}

	if !filter.SessionStartedAfter.IsZero() {
		queryString.WriteString(templateWithSessionStartSuffix)
		operands = append(operands, filter.SessionStartedAfter)
	}

	if filter.HostIDGreaterThan != nil {
		queryString.WriteString(templateWithHostIDGreaterSuffix)
		operands = append(operands, filter.HostIDGreaterThan)
	}

	queryString.WriteString(templateWithOrderBySessionStartSuffix)

	if filter.MaxRecordCount > 0 {
		queryString.WriteString(templateWithLimitSuffix)
		operands = append(operands, filter.MaxRecordCount)
	}

	compiledQryString := queryString.String()

	var rows []sqlplugin.ClusterMembershipRow
	if err := mdb.SelectContext(ctx,
		&rows,
		compiledQryString,
		operands...,
	); err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].SessionStart = mdb.converter.FromMySQLDateTime(rows[i].SessionStart)
		rows[i].LastHeartbeat = mdb.converter.FromMySQLDateTime(rows[i].LastHeartbeat)
		rows[i].RecordExpiry = mdb.converter.FromMySQLDateTime(rows[i].RecordExpiry)
	}
	return rows, nil
}

func (mdb *db) PruneClusterMembership(
	ctx context.Context,
	filter *sqlplugin.PruneClusterMembershipFilter,
) (sql.Result, error) {
	return mdb.ExecContext(ctx,
		templatePruneStaleClusterMembership,
		constMembershipPartition,
		mdb.converter.ToMySQLDateTime(filter.PruneRecordsBefore),
	)
}
