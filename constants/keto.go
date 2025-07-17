package constants

const (
	// Keto API Operation Names
	OperationCreateRelationship         = "RelationshipApiService.CreateRelationship"
	OperationDeleteRelationships        = "RelationshipApiService.DeleteRelationships"
	OperationPatchRelationships         = "RelationshipApiService.PatchRelationships"
	OperationGetRelationships           = "RelationshipApiService.GetRelationships"
	OperationListRelationshipNamespaces = "RelationshipApiService.ListRelationshipNamespaces"
	OperationCheckPermission            = "PermissionApiService.CheckPermission"
	OperationCheckPermissionOrError     = "PermissionApiService.CheckPermissionOrError"
	OperationPostCheckPermission        = "PermissionApiService.PostCheckPermission"
	OperationPostCheckPermissionOrError = "PermissionApiService.PostCheckPermissionOrError"
	OperationExpandPermissions          = "PermissionApiService.ExpandPermissions"
	OperationCheckOplSyntax             = "RelationshipApiService.CheckOplSyntax"
	OperationGetVersion                 = "MetadataApiService.GetVersion"
	OperationIsAlive                    = "MetadataApiService.IsAlive"
	OperationIsReady                    = "MetadataApiService.IsReady"

	// Keto API Descriptions
	KetoWriteApiDescription = "Keto Write API"
	KetoReadApiDescription  = "Keto Read API"

	// Endpoint paths
	BatchPermissionCheckEndpoint = "/check/permission/bulk"
)
