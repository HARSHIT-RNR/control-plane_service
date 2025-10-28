package grpc_server

import (
	"context"
	"errors"

	pb "cp_service/api/proto/pb"
	"cp_service/internal/adapters/database/db"
	"cp_service/internal/adapters/helpers"
	"cp_service/internal/core/services"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrganizationHandler struct {
	pb.UnimplementedOrganizationServiceServer
	orgService *services.OrganizationService
}

func NewOrganizationHandler(orgService *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{orgService: orgService}
}

// Department handlers

func (h *OrganizationHandler) CreateDepartment(ctx context.Context, req *pb.CreateDepartmentRequest) (*pb.DepartmentResponse, error) {
	// Generate new UUID for department
	newID := uuid.New()
	
	// Parse tenant_id from request
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}
	
	params := db.CreateDepartmentParams{
		ID:       helpers.UUIDToPgUUID(newID),
		Name:     req.Name,
		TenantID: helpers.UUIDToPgUUID(tenantID),
	}

	if req.Description != "" {
		params.Description = helpers.StringToPgText(req.Description)
	}

	dept, err := h.orgService.CreateDepartment(ctx, params)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.DepartmentResponse{Department: dbDepartmentToProto(&dept)}, nil
}

func (h *OrganizationHandler) GetDepartment(ctx context.Context, req *pb.GetDepartmentRequest) (*pb.DepartmentResponse, error) {
	dept, err := h.orgService.GetDepartment(ctx, req.Id)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.DepartmentResponse{Department: dbDepartmentToProto(&dept)}, nil
}

func (h *OrganizationHandler) ListDepartments(ctx context.Context, req *pb.ListDepartmentsRequest) (*pb.ListDepartmentsResponse, error) {
	// Extract tenant_id from auth context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, helpers.ToGRPCError(errors.New("tenant_id not found in context"))
	}
	
	depts, err := h.orgService.ListDepartments(ctx, tenantID.String())
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	pbDepts := make([]*pb.Department, len(depts))
	for i, dept := range depts {
		pbDepts[i] = dbDepartmentToProto(&dept)
	}

	return &pb.ListDepartmentsResponse{Departments: pbDepts}, nil
}

func (h *OrganizationHandler) UpdateDepartment(ctx context.Context, req *pb.UpdateDepartmentRequest) (*pb.DepartmentResponse, error) {
	deptID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}
	
	params := db.UpdateDepartmentParams{
		ID:   helpers.UUIDToPgUUID(deptID),
		Name: req.Name,
	}

	if req.Description != "" {
		params.Description = helpers.StringToPgText(req.Description)
	}

	dept, err := h.orgService.UpdateDepartment(ctx, params)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.DepartmentResponse{Department: dbDepartmentToProto(&dept)}, nil
}

func (h *OrganizationHandler) DeleteDepartment(ctx context.Context, req *pb.DeleteDepartmentRequest) (*pb.DeleteDepartmentResponse, error) {
	if err := h.orgService.DeleteDepartment(ctx, req.Id); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.DeleteDepartmentResponse{Success: true}, nil
}

// Designation handlers

func (h *OrganizationHandler) CreateDesignation(ctx context.Context, req *pb.CreateDesignationRequest) (*pb.DesignationResponse, error) {
	// Generate new UUID for designation
	newID := uuid.New()
	
	// Parse tenant_id from request
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}
	
	params := db.CreateDesignationParams{
		ID:       helpers.UUIDToPgUUID(newID),
		Name:     req.Name,
		TenantID: helpers.UUIDToPgUUID(tenantID),
	}

	if req.Description != "" {
		params.Description = helpers.StringToPgText(req.Description)
	}

	desig, err := h.orgService.CreateDesignation(ctx, params)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.DesignationResponse{Designation: dbDesignationToProto(&desig)}, nil
}

func (h *OrganizationHandler) GetDesignation(ctx context.Context, req *pb.GetDesignationRequest) (*pb.DesignationResponse, error) {
	desig, err := h.orgService.GetDesignation(ctx, req.Id)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.DesignationResponse{Designation: dbDesignationToProto(&desig)}, nil
}

func (h *OrganizationHandler) ListDesignations(ctx context.Context, req *pb.ListDesignationsRequest) (*pb.ListDesignationsResponse, error) {
	// Extract tenant_id from auth context
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return nil, helpers.ToGRPCError(errors.New("tenant_id not found in context"))
	}
	
	desigs, err := h.orgService.ListDesignations(ctx, tenantID.String())
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	pbDesigs := make([]*pb.Designation, len(desigs))
	for i, desig := range desigs {
		pbDesigs[i] = dbDesignationToProto(&desig)
	}

	return &pb.ListDesignationsResponse{Designations: pbDesigs}, nil
}

func (h *OrganizationHandler) UpdateDesignation(ctx context.Context, req *pb.UpdateDesignationRequest) (*pb.DesignationResponse, error) {
	desigID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}
	
	params := db.UpdateDesignationParams{
		ID:   helpers.UUIDToPgUUID(desigID),
		Name: req.Name,
	}

	if req.Description != "" {
		params.Description = helpers.StringToPgText(req.Description)
	}

	desig, err := h.orgService.UpdateDesignation(ctx, params)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.DesignationResponse{Designation: dbDesignationToProto(&desig)}, nil
}

func (h *OrganizationHandler) DeleteDesignation(ctx context.Context, req *pb.DeleteDesignationRequest) (*pb.DeleteDesignationResponse, error) {
	if err := h.orgService.DeleteDesignation(ctx, req.Id); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.DeleteDesignationResponse{Success: true}, nil
}

// Helper functions

func dbDepartmentToProto(dept *db.Department) *pb.Department {
	pbDept := &pb.Department{
		Id:       helpers.PgUUIDToString(dept.ID),
		TenantId: helpers.PgUUIDToString(dept.TenantID),
		Name:     dept.Name,
	}

	if dept.Description.Valid {
		pbDept.Description = dept.Description.String
	}
	if dept.CreatedAt.Valid {
		pbDept.CreatedAt = timestamppb.New(dept.CreatedAt.Time)
	}
	if dept.UpdatedAt.Valid {
		pbDept.UpdatedAt = timestamppb.New(dept.UpdatedAt.Time)
	}

	return pbDept
}

func dbDesignationToProto(desig *db.Designation) *pb.Designation {
	pbDesig := &pb.Designation{
		Id:       helpers.PgUUIDToString(desig.ID),
		TenantId: helpers.PgUUIDToString(desig.TenantID),
		Name:     desig.Name,
	}

	if desig.Description.Valid {
		pbDesig.Description = desig.Description.String
	}
	if desig.CreatedAt.Valid {
		pbDesig.CreatedAt = timestamppb.New(desig.CreatedAt.Time)
	}
	if desig.UpdatedAt.Valid {
		pbDesig.UpdatedAt = timestamppb.New(desig.UpdatedAt.Time)
	}

	return pbDesig
}
