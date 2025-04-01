package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	pb "github.com/scalar-labs/terraform-provider-scalardb/proto/scalardb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port = flag.Int("port", 60051, "The server port")
)

// mockServer is used to implement ScalarDB Cluster gRPC API.
type mockServer struct {
	pb.UnimplementedDistributedTransactionAdminServer
	namespaces map[string]bool
	tables     map[string]map[string]*pb.TableMetadata
}

func newMockServer() *mockServer {
	return &mockServer{
		namespaces: make(map[string]bool),
		tables:     make(map[string]map[string]*pb.TableMetadata),
	}
}

// CreateNamespace implements the CreateNamespace RPC.
func (s *mockServer) CreateNamespace(ctx context.Context, req *pb.CreateNamespaceRequest) (*pb.CreateNamespaceResponse, error) {
	log.Printf("CreateNamespace: %v", req)
	if _, exists := s.namespaces[req.NamespaceName]; exists && !req.IfNotExists {
		return nil, fmt.Errorf("namespace %s already exists", req.NamespaceName)
	}
	s.namespaces[req.NamespaceName] = true
	if _, exists := s.tables[req.NamespaceName]; !exists {
		s.tables[req.NamespaceName] = make(map[string]*pb.TableMetadata)
	}
	return &pb.CreateNamespaceResponse{}, nil
}

// DropNamespace implements the DropNamespace RPC.
func (s *mockServer) DropNamespace(ctx context.Context, req *pb.DropNamespaceRequest) (*pb.DropNamespaceResponse, error) {
	log.Printf("DropNamespace: %v", req)
	if _, exists := s.namespaces[req.NamespaceName]; !exists && !req.IfExists {
		return nil, fmt.Errorf("namespace %s does not exist", req.NamespaceName)
	}
	delete(s.namespaces, req.NamespaceName)
	delete(s.tables, req.NamespaceName)
	return &pb.DropNamespaceResponse{}, nil
}

// NamespaceExists implements the NamespaceExists RPC.
func (s *mockServer) NamespaceExists(ctx context.Context, req *pb.NamespaceExistsRequest) (*pb.NamespaceExistsResponse, error) {
	log.Printf("NamespaceExists: %v", req)
	exists := false
	if _, ok := s.namespaces[req.NamespaceName]; ok {
		exists = true
	}
	return &pb.NamespaceExistsResponse{
		Exists: exists,
	}, nil
}

// CreateTable implements the CreateTable RPC.
func (s *mockServer) CreateTable(ctx context.Context, req *pb.CreateTableRequest) (*pb.CreateTableResponse, error) {
	log.Printf("CreateTable: %v", req)
	if _, exists := s.namespaces[req.NamespaceName]; !exists {
		return nil, fmt.Errorf("namespace %s does not exist", req.NamespaceName)
	}
	if _, exists := s.tables[req.NamespaceName][req.TableName]; exists && !req.IfNotExists {
		return nil, fmt.Errorf("table %s.%s already exists", req.NamespaceName, req.TableName)
	}
	s.tables[req.NamespaceName][req.TableName] = req.TableMetadata
	return &pb.CreateTableResponse{}, nil
}

// DropTable implements the DropTable RPC.
func (s *mockServer) DropTable(ctx context.Context, req *pb.DropTableRequest) (*pb.DropTableResponse, error) {
	log.Printf("DropTable: %v", req)
	if _, exists := s.namespaces[req.NamespaceName]; !exists {
		return nil, fmt.Errorf("namespace %s does not exist", req.NamespaceName)
	}
	if _, exists := s.tables[req.NamespaceName][req.TableName]; !exists && !req.IfExists {
		return nil, fmt.Errorf("table %s.%s does not exist", req.NamespaceName, req.TableName)
	}
	delete(s.tables[req.NamespaceName], req.TableName)
	return &pb.DropTableResponse{}, nil
}

// TableExists implements the TableExists RPC.
func (s *mockServer) TableExists(ctx context.Context, req *pb.TableExistsRequest) (*pb.TableExistsResponse, error) {
	log.Printf("TableExists: %v", req)
	exists := false
	if _, ok := s.namespaces[req.NamespaceName]; ok {
		if _, ok := s.tables[req.NamespaceName][req.TableName]; ok {
			exists = true
		}
	}
	return &pb.TableExistsResponse{
		Exists: exists,
	}, nil
}

// GetTableMetadata implements the GetTableMetadata RPC.
func (s *mockServer) GetTableMetadata(ctx context.Context, req *pb.GetTableMetadataRequest) (*pb.GetTableMetadataResponse, error) {
	log.Printf("GetTableMetadata: %v", req)
	if _, exists := s.namespaces[req.NamespaceName]; !exists {
		return nil, fmt.Errorf("namespace %s does not exist", req.NamespaceName)
	}
	if _, exists := s.tables[req.NamespaceName][req.TableName]; !exists {
		return nil, fmt.Errorf("table %s.%s does not exist", req.NamespaceName, req.TableName)
	}
	return &pb.GetTableMetadataResponse{
		TableMetadata: s.tables[req.NamespaceName][req.TableName],
	}, nil
}

// Implement other required methods from the DistributedTransactionAdmin interface
// with empty implementations to satisfy the interface.

func (s *mockServer) TruncateTable(ctx context.Context, req *pb.TruncateTableRequest) (*pb.TruncateTableResponse, error) {
	return &pb.TruncateTableResponse{}, nil
}

func (s *mockServer) CreateIndex(ctx context.Context, req *pb.CreateIndexRequest) (*pb.CreateIndexResponse, error) {
	return &pb.CreateIndexResponse{}, nil
}

func (s *mockServer) DropIndex(ctx context.Context, req *pb.DropIndexRequest) (*pb.DropIndexResponse, error) {
	return &pb.DropIndexResponse{}, nil
}

func (s *mockServer) IndexExists(ctx context.Context, req *pb.IndexExistsRequest) (*pb.IndexExistsResponse, error) {
	return &pb.IndexExistsResponse{Exists: false}, nil
}

func (s *mockServer) RepairNamespace(ctx context.Context, req *pb.RepairNamespaceRequest) (*pb.RepairNamespaceResponse, error) {
	return &pb.RepairNamespaceResponse{}, nil
}

func (s *mockServer) RepairTable(ctx context.Context, req *pb.RepairTableRequest) (*pb.RepairTableResponse, error) {
	return &pb.RepairTableResponse{}, nil
}

func (s *mockServer) AddNewColumnToTable(ctx context.Context, req *pb.AddNewColumnToTableRequest) (*pb.AddNewColumnToTableResponse, error) {
	return &pb.AddNewColumnToTableResponse{}, nil
}

func (s *mockServer) CreateCoordinatorTables(ctx context.Context, req *pb.CreateCoordinatorTablesRequest) (*pb.CreateCoordinatorTablesResponse, error) {
	return &pb.CreateCoordinatorTablesResponse{}, nil
}

func (s *mockServer) DropCoordinatorTables(ctx context.Context, req *pb.DropCoordinatorTablesRequest) (*pb.DropCoordinatorTablesResponse, error) {
	return &pb.DropCoordinatorTablesResponse{}, nil
}

func (s *mockServer) TruncateCoordinatorTables(ctx context.Context, req *pb.TruncateCoordinatorTablesRequest) (*pb.TruncateCoordinatorTablesResponse, error) {
	return &pb.TruncateCoordinatorTablesResponse{}, nil
}

func (s *mockServer) CoordinatorTablesExist(ctx context.Context, req *pb.CoordinatorTablesExistRequest) (*pb.CoordinatorTablesExistResponse, error) {
	return &pb.CoordinatorTablesExistResponse{Exist: false}, nil
}

func (s *mockServer) RepairCoordinatorTables(ctx context.Context, req *pb.RepairCoordinatorTablesRequest) (*pb.RepairCoordinatorTablesResponse, error) {
	return &pb.RepairCoordinatorTablesResponse{}, nil
}

func (s *mockServer) GetNamespaceNames(ctx context.Context, req *pb.GetNamespaceNamesRequest) (*pb.GetNamespaceNamesResponse, error) {
	namespaceNames := make([]string, 0, len(s.namespaces))
	for name := range s.namespaces {
		namespaceNames = append(namespaceNames, name)
	}
	return &pb.GetNamespaceNamesResponse{NamespaceNames: namespaceNames}, nil
}

func (s *mockServer) GetNamespaceTableNames(ctx context.Context, req *pb.GetNamespaceTableNamesRequest) (*pb.GetNamespaceTableNamesResponse, error) {
	if _, exists := s.namespaces[req.NamespaceName]; !exists {
		return nil, fmt.Errorf("namespace %s does not exist", req.NamespaceName)
	}
	tableNames := make([]string, 0, len(s.tables[req.NamespaceName]))
	for name := range s.tables[req.NamespaceName] {
		tableNames = append(tableNames, name)
	}
	return &pb.GetNamespaceTableNamesResponse{TableNames: tableNames}, nil
}

func (s *mockServer) ImportTable(ctx context.Context, req *pb.ImportTableRequest) (*pb.ImportTableResponse, error) {
	return &pb.ImportTableResponse{}, nil
}

func (s *mockServer) Upgrade(ctx context.Context, req *pb.UpgradeRequest) (*pb.UpgradeResponse, error) {
	return &pb.UpgradeResponse{}, nil
}

func (s *mockServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return &pb.CreateUserResponse{}, nil
}

func (s *mockServer) AlterUser(ctx context.Context, req *pb.AlterUserRequest) (*pb.AlterUserResponse, error) {
	return &pb.AlterUserResponse{}, nil
}

func (s *mockServer) DropUser(ctx context.Context, req *pb.DropUserRequest) (*pb.DropUserResponse, error) {
	return &pb.DropUserResponse{}, nil
}

func (s *mockServer) Grant(ctx context.Context, req *pb.GrantRequest) (*pb.GrantResponse, error) {
	return &pb.GrantResponse{}, nil
}

func (s *mockServer) Revoke(ctx context.Context, req *pb.RevokeRequest) (*pb.RevokeResponse, error) {
	return &pb.RevokeResponse{}, nil
}

func (s *mockServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	return &pb.GetUserResponse{}, nil
}

func (s *mockServer) GetUsers(ctx context.Context, req *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {
	return &pb.GetUsersResponse{}, nil
}

func (s *mockServer) GetCurrentUser(ctx context.Context, req *pb.GetCurrentUserRequest) (*pb.GetCurrentUserResponse, error) {
	return &pb.GetCurrentUserResponse{}, nil
}

func (s *mockServer) GetPrivileges(ctx context.Context, req *pb.GetPrivilegesRequest) (*pb.GetPrivilegesResponse, error) {
	return &pb.GetPrivilegesResponse{}, nil
}

func (s *mockServer) CreatePolicy(ctx context.Context, req *pb.CreatePolicyRequest) (*pb.CreatePolicyResponse, error) {
	return &pb.CreatePolicyResponse{}, nil
}

func (s *mockServer) EnablePolicy(ctx context.Context, req *pb.EnablePolicyRequest) (*pb.EnablePolicyResponse, error) {
	return &pb.EnablePolicyResponse{}, nil
}

func (s *mockServer) DisablePolicy(ctx context.Context, req *pb.DisablePolicyRequest) (*pb.DisablePolicyResponse, error) {
	return &pb.DisablePolicyResponse{}, nil
}

func (s *mockServer) GetPolicy(ctx context.Context, req *pb.GetPolicyRequest) (*pb.GetPolicyResponse, error) {
	return &pb.GetPolicyResponse{}, nil
}

func (s *mockServer) GetPolicies(ctx context.Context, req *pb.GetPoliciesRequest) (*pb.GetPoliciesResponse, error) {
	return &pb.GetPoliciesResponse{}, nil
}

func (s *mockServer) CreateLevel(ctx context.Context, req *pb.CreateLevelRequest) (*pb.CreateLevelResponse, error) {
	return &pb.CreateLevelResponse{}, nil
}

func (s *mockServer) DropLevel(ctx context.Context, req *pb.DropLevelRequest) (*pb.DropLevelResponse, error) {
	return &pb.DropLevelResponse{}, nil
}

func (s *mockServer) GetLevel(ctx context.Context, req *pb.GetLevelRequest) (*pb.GetLevelResponse, error) {
	return &pb.GetLevelResponse{}, nil
}

func (s *mockServer) GetLevels(ctx context.Context, req *pb.GetLevelsRequest) (*pb.GetLevelsResponse, error) {
	return &pb.GetLevelsResponse{}, nil
}

func (s *mockServer) CreateCompartment(ctx context.Context, req *pb.CreateCompartmentRequest) (*pb.CreateCompartmentResponse, error) {
	return &pb.CreateCompartmentResponse{}, nil
}

func (s *mockServer) DropCompartment(ctx context.Context, req *pb.DropCompartmentRequest) (*pb.DropCompartmentResponse, error) {
	return &pb.DropCompartmentResponse{}, nil
}

func (s *mockServer) GetCompartment(ctx context.Context, req *pb.GetCompartmentRequest) (*pb.GetCompartmentResponse, error) {
	return &pb.GetCompartmentResponse{}, nil
}

func (s *mockServer) GetCompartments(ctx context.Context, req *pb.GetCompartmentsRequest) (*pb.GetCompartmentsResponse, error) {
	return &pb.GetCompartmentsResponse{}, nil
}

func (s *mockServer) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error) {
	return &pb.CreateGroupResponse{}, nil
}

func (s *mockServer) DropGroup(ctx context.Context, req *pb.DropGroupRequest) (*pb.DropGroupResponse, error) {
	return &pb.DropGroupResponse{}, nil
}

func (s *mockServer) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupResponse, error) {
	return &pb.GetGroupResponse{}, nil
}

func (s *mockServer) GetGroups(ctx context.Context, req *pb.GetGroupsRequest) (*pb.GetGroupsResponse, error) {
	return &pb.GetGroupsResponse{}, nil
}

func (s *mockServer) SetLevelsToUser(ctx context.Context, req *pb.SetLevelsToUserRequest) (*pb.SetLevelsToUserResponse, error) {
	return &pb.SetLevelsToUserResponse{}, nil
}

func (s *mockServer) AddCompartmentToUser(ctx context.Context, req *pb.AddCompartmentToUserRequest) (*pb.AddCompartmentToUserResponse, error) {
	return &pb.AddCompartmentToUserResponse{}, nil
}

func (s *mockServer) RemoveCompartmentFromUser(ctx context.Context, req *pb.RemoveCompartmentFromUserRequest) (*pb.RemoveCompartmentFromUserResponse, error) {
	return &pb.RemoveCompartmentFromUserResponse{}, nil
}

func (s *mockServer) AddGroupToUser(ctx context.Context, req *pb.AddGroupToUserRequest) (*pb.AddGroupToUserResponse, error) {
	return &pb.AddGroupToUserResponse{}, nil
}

func (s *mockServer) RemoveGroupFromUser(ctx context.Context, req *pb.RemoveGroupFromUserRequest) (*pb.RemoveGroupFromUserResponse, error) {
	return &pb.RemoveGroupFromUserResponse{}, nil
}

func (s *mockServer) DropUserTagInfoFromUser(ctx context.Context, req *pb.DropUserTagInfoFromUserRequest) (*pb.DropUserTagInfoFromUserResponse, error) {
	return &pb.DropUserTagInfoFromUserResponse{}, nil
}

func (s *mockServer) GetUserTagInfo(ctx context.Context, req *pb.GetUserTagInfoRequest) (*pb.GetUserTagInfoResponse, error) {
	return &pb.GetUserTagInfoResponse{}, nil
}

func (s *mockServer) CreateNamespacePolicy(ctx context.Context, req *pb.CreateNamespacePolicyRequest) (*pb.CreateNamespacePolicyResponse, error) {
	return &pb.CreateNamespacePolicyResponse{}, nil
}

func (s *mockServer) EnableNamespacePolicy(ctx context.Context, req *pb.EnableNamespacePolicyRequest) (*pb.EnableNamespacePolicyResponse, error) {
	return &pb.EnableNamespacePolicyResponse{}, nil
}

func (s *mockServer) DisableNamespacePolicy(ctx context.Context, req *pb.DisableNamespacePolicyRequest) (*pb.DisableNamespacePolicyResponse, error) {
	return &pb.DisableNamespacePolicyResponse{}, nil
}

func (s *mockServer) GetNamespacePolicy(ctx context.Context, req *pb.GetNamespacePolicyRequest) (*pb.GetNamespacePolicyResponse, error) {
	return &pb.GetNamespacePolicyResponse{}, nil
}

func (s *mockServer) GetNamespacePolicies(ctx context.Context, req *pb.GetNamespacePoliciesRequest) (*pb.GetNamespacePoliciesResponse, error) {
	return &pb.GetNamespacePoliciesResponse{}, nil
}

func (s *mockServer) CreateTablePolicy(ctx context.Context, req *pb.CreateTablePolicyRequest) (*pb.CreateTablePolicyResponse, error) {
	return &pb.CreateTablePolicyResponse{}, nil
}

func (s *mockServer) EnableTablePolicy(ctx context.Context, req *pb.EnableTablePolicyRequest) (*pb.EnableTablePolicyResponse, error) {
	return &pb.EnableTablePolicyResponse{}, nil
}

func (s *mockServer) DisableTablePolicy(ctx context.Context, req *pb.DisableTablePolicyRequest) (*pb.DisableTablePolicyResponse, error) {
	return &pb.DisableTablePolicyResponse{}, nil
}

func (s *mockServer) GetTablePolicy(ctx context.Context, req *pb.GetTablePolicyRequest) (*pb.GetTablePolicyResponse, error) {
	return &pb.GetTablePolicyResponse{}, nil
}

func (s *mockServer) GetTablePolicies(ctx context.Context, req *pb.GetTablePoliciesRequest) (*pb.GetTablePoliciesResponse, error) {
	return &pb.GetTablePoliciesResponse{}, nil
}

func main() {
	flag.Parse()

	// Check if port is specified via environment variable
	if portEnv := os.Getenv("SCALARDB_MOCK_PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			*port = p
		}
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterDistributedTransactionAdminServer(s, newMockServer())
	// Register reflection service on gRPC server.
	reflection.Register(s)
	log.Printf("Mock ScalarDB Cluster server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
