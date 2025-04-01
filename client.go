package main

import (
	"context"
	"fmt"
	"strconv"

	pb "github.com/scalar-labs/terraform-provider-scalardb/proto/scalardb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a client for the ScalarDB API.
type Client struct {
	Host     string
	Port     int
	Username string
	Password string
	conn     *grpc.ClientConn
	admin    pb.DistributedTransactionAdminClient
}

// NewClient creates a new ScalarDB client.
func NewClient(host string, port int, username, password string) *Client {
	return &Client{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
}

// Connect establishes a connection to the ScalarDB server.
func (c *Client) Connect() error {
	address := fmt.Sprintf("%s:%d", c.Host, c.Port)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to ScalarDB server: %w", err)
	}
	c.conn = conn
	c.admin = pb.NewDistributedTransactionAdminClient(conn)
	return nil
}

// Close closes the connection to the ScalarDB server.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// getRequestHeader creates a request header with authentication if credentials are provided.
func (c *Client) getRequestHeader() *pb.RequestHeader {
	header := &pb.RequestHeader{
		HopLimit: 10, // Default hop limit
	}
	if c.Username != "" && c.Password != "" {
		// In a real implementation, this would be a proper auth token
		// For now, we just concatenate username and password
		header.AuthToken = &c.Username
	}
	return header
}

// CreateNamespace creates a new namespace in ScalarDB.
func (c *Client) CreateNamespace(ctx context.Context, name string, options map[string]interface{}) error {
	if c.admin == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// Convert options to string map
	strOptions := make(map[string]string)
	for k, v := range options {
		switch val := v.(type) {
		case string:
			strOptions[k] = val
		case int:
			strOptions[k] = strconv.Itoa(val)
		case bool:
			strOptions[k] = strconv.FormatBool(val)
		default:
			strOptions[k] = fmt.Sprintf("%v", val)
		}
	}

	req := &pb.CreateNamespaceRequest{
		RequestHeader:  c.getRequestHeader(),
		NamespaceName:  name,
		Options:        strOptions,
		IfNotExists:    true,
	}

	_, err := c.admin.CreateNamespace(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	return nil
}

// DeleteNamespace deletes a namespace from ScalarDB.
func (c *Client) DeleteNamespace(ctx context.Context, name string) error {
	if c.admin == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	req := &pb.DropNamespaceRequest{
		RequestHeader:  c.getRequestHeader(),
		NamespaceName:  name,
		IfExists:       true,
	}

	_, err := c.admin.DropNamespace(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	return nil
}

// NamespaceExists checks if a namespace exists in ScalarDB.
func (c *Client) NamespaceExists(ctx context.Context, name string) (bool, error) {
	if c.admin == nil {
		if err := c.Connect(); err != nil {
			return false, err
		}
	}

	req := &pb.NamespaceExistsRequest{
		RequestHeader:  c.getRequestHeader(),
		NamespaceName:  name,
	}

	resp, err := c.admin.NamespaceExists(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to check if namespace exists: %w", err)
	}

	return resp.Exists, nil
}

// convertDataType converts a string data type to a pb.DataType.
func convertDataType(dataType string) pb.DataType {
	switch dataType {
	case "BOOLEAN":
		return pb.DataType_DATA_TYPE_BOOLEAN
	case "INT":
		return pb.DataType_DATA_TYPE_INT
	case "BIGINT":
		return pb.DataType_DATA_TYPE_BIGINT
	case "FLOAT":
		return pb.DataType_DATA_TYPE_FLOAT
	case "DOUBLE":
		return pb.DataType_DATA_TYPE_DOUBLE
	case "TEXT":
		return pb.DataType_DATA_TYPE_TEXT
	case "BLOB":
		return pb.DataType_DATA_TYPE_BLOB
	case "DATE":
		return pb.DataType_DATA_TYPE_DATE
	case "TIME":
		return pb.DataType_DATA_TYPE_TIME
	case "TIMESTAMP":
		return pb.DataType_DATA_TYPE_TIMESTAMP
	case "TIMESTAMPTZ":
		return pb.DataType_DATA_TYPE_TIMESTAMPTZ
	default:
		return pb.DataType_DATA_TYPE_TEXT // Default to TEXT
	}
}

// convertClusteringOrder converts a string clustering order to a pb.ClusteringOrder.
func convertClusteringOrder(order string) pb.ClusteringOrder {
	switch order {
	case "ASC":
		return pb.ClusteringOrder_CLUSTERING_ORDER_ASC
	case "DESC":
		return pb.ClusteringOrder_CLUSTERING_ORDER_DESC
	default:
		return pb.ClusteringOrder_CLUSTERING_ORDER_ASC // Default to ASC
	}
}

// CreateTable creates a new table in ScalarDB.
func (c *Client) CreateTable(ctx context.Context, namespace, name string, columns map[string]map[string]interface{}, options map[string]interface{}) error {
	if c.admin == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// Convert options to string map
	strOptions := make(map[string]string)
	for k, v := range options {
		switch val := v.(type) {
		case string:
			strOptions[k] = val
		case int:
			strOptions[k] = strconv.Itoa(val)
		case bool:
			strOptions[k] = strconv.FormatBool(val)
		case map[string]interface{}:
			// Handle clustering_order map
			if k == "clustering_order" {
				for ck, cv := range val {
					strOptions[fmt.Sprintf("clustering_order.%s", ck)] = fmt.Sprintf("%v", cv)
				}
			}
		default:
			strOptions[k] = fmt.Sprintf("%v", val)
		}
	}

	// Build table metadata
	tableMetadata := &pb.TableMetadata{
		Columns:                 make(map[string]pb.DataType),
		PartitionKeyColumnNames: []string{},
		ClusteringKeyColumnNames: []string{},
		ClusteringOrders:        make(map[string]pb.ClusteringOrder),
		SecondaryIndexColumnNames: []string{},
		EncryptedColumns:        []string{},
	}

	// Process columns
	for colName, colProps := range columns {
		// Add column to columns map
		if dataTypeStr, ok := colProps["type"].(string); ok {
			tableMetadata.Columns[colName] = convertDataType(dataTypeStr)
		}

		// Add to partition key if specified
		if isPartitionKey, ok := colProps["partition_key"].(bool); ok && isPartitionKey {
			tableMetadata.PartitionKeyColumnNames = append(tableMetadata.PartitionKeyColumnNames, colName)
		}

		// Add to clustering key if specified
		if isClusteringKey, ok := colProps["clustering_key"].(bool); ok && isClusteringKey {
			tableMetadata.ClusteringKeyColumnNames = append(tableMetadata.ClusteringKeyColumnNames, colName)
		}

		// Add to secondary index if specified
		if isSecondaryIndex, ok := colProps["secondary_index"].(bool); ok && isSecondaryIndex {
			tableMetadata.SecondaryIndexColumnNames = append(tableMetadata.SecondaryIndexColumnNames, colName)
		}

		// Add to encrypted columns if specified
		if isEncrypted, ok := colProps["encrypted"].(bool); ok && isEncrypted {
			tableMetadata.EncryptedColumns = append(tableMetadata.EncryptedColumns, colName)
		}
	}

	// Add clustering orders
	if clusteringOrderMap, ok := options["clustering_order"].(map[string]interface{}); ok {
		for colName, orderVal := range clusteringOrderMap {
			if orderStr, ok := orderVal.(string); ok {
				tableMetadata.ClusteringOrders[colName] = convertClusteringOrder(orderStr)
			}
		}
	}

	req := &pb.CreateTableRequest{
		RequestHeader:  c.getRequestHeader(),
		NamespaceName:  namespace,
		TableName:      name,
		TableMetadata:  tableMetadata,
		Options:        strOptions,
		IfNotExists:    true,
	}

	_, err := c.admin.CreateTable(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// DeleteTable deletes a table from ScalarDB.
func (c *Client) DeleteTable(ctx context.Context, namespace, name string) error {
	if c.admin == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	req := &pb.DropTableRequest{
		RequestHeader:  c.getRequestHeader(),
		NamespaceName:  namespace,
		TableName:      name,
		IfExists:       true,
	}

	_, err := c.admin.DropTable(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete table: %w", err)
	}

	return nil
}

// TableExists checks if a table exists in ScalarDB.
func (c *Client) TableExists(ctx context.Context, namespace, name string) (bool, error) {
	if c.admin == nil {
		if err := c.Connect(); err != nil {
			return false, err
		}
	}

	req := &pb.TableExistsRequest{
		RequestHeader:  c.getRequestHeader(),
		NamespaceName:  namespace,
		TableName:      name,
	}

	resp, err := c.admin.TableExists(ctx, req)
	if err != nil {
		return false, fmt.Errorf("failed to check if table exists: %w", err)
	}

	return resp.Exists, nil
}

// GetTableSchema gets the schema of a table from ScalarDB.
func (c *Client) GetTableSchema(ctx context.Context, namespace, name string) (map[string]map[string]interface{}, map[string]interface{}, error) {
	if c.admin == nil {
		if err := c.Connect(); err != nil {
			return nil, nil, err
		}
	}

	req := &pb.GetTableMetadataRequest{
		RequestHeader:  c.getRequestHeader(),
		NamespaceName:  namespace,
		TableName:      name,
	}

	resp, err := c.admin.GetTableMetadata(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get table metadata: %w", err)
	}

	if resp.TableMetadata == nil {
		return nil, nil, fmt.Errorf("table metadata not found")
	}

	// Convert pb.TableMetadata to map[string]map[string]interface{}
	columns := make(map[string]map[string]interface{})
	for colName, dataType := range resp.TableMetadata.Columns {
		colProps := make(map[string]interface{})
		colProps["type"] = dataType.String()

		// Check if column is a partition key
		for _, pkCol := range resp.TableMetadata.PartitionKeyColumnNames {
			if pkCol == colName {
				colProps["partition_key"] = true
				break
			}
		}

		// Check if column is a clustering key
		for _, ckCol := range resp.TableMetadata.ClusteringKeyColumnNames {
			if ckCol == colName {
				colProps["clustering_key"] = true
				break
			}
		}

		// Check if column is a secondary index
		for _, idxCol := range resp.TableMetadata.SecondaryIndexColumnNames {
			if idxCol == colName {
				colProps["secondary_index"] = true
				break
			}
		}

		// Check if column is encrypted
		for _, encCol := range resp.TableMetadata.EncryptedColumns {
			if encCol == colName {
				colProps["encrypted"] = true
				break
			}
		}

		columns[colName] = colProps
	}

	// Convert clustering orders to options
	options := make(map[string]interface{})
	if len(resp.TableMetadata.ClusteringOrders) > 0 {
		clusteringOrder := make(map[string]interface{})
		for colName, order := range resp.TableMetadata.ClusteringOrders {
			clusteringOrder[colName] = order.String()
		}
		options["clustering_order"] = clusteringOrder
	}

	return columns, options, nil
}
