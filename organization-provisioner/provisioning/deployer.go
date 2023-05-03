package provisioning

import (
	"github.com/pulumi/pulumi-keycloak/sdk/v5/go/keycloak"
	"github.com/pulumi/pulumi-keycloak/sdk/v5/go/keycloak/openid"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func deployer(orgId string, orgName string) pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		realm, err := keycloak.NewRealm(ctx, "realm", &keycloak.RealmArgs{
			Realm:       pulumi.String(orgId),
			DisplayName: pulumi.String(orgName),
		})

		if err != nil {
			return err
		}

		memberRole, err := keycloak.NewRole(ctx, "member-role", &keycloak.RoleArgs{
			Name:    pulumi.StringPtr("member"),
			RealmId: realm.ID(),
		})

		if err != nil {
			return err
		}

		adminRole, err := keycloak.NewRole(ctx, "admin-role", &keycloak.RoleArgs{
			Name:           pulumi.StringPtr("admin"),
			RealmId:        realm.ID(),
			CompositeRoles: pulumi.StringArray{memberRole.ID()},
		})

		if err != nil {
			return err
		}

		ownerRole, err := keycloak.NewRole(ctx, "owner-role", &keycloak.RoleArgs{
			Name:           pulumi.StringPtr("owner"),
			RealmId:        realm.ID(),
			CompositeRoles: pulumi.StringArray{adminRole.ID()},
		})

		if err != nil {
			return err
		}

		_, err = openid.NewClient(ctx, "frontend-client", &openid.ClientArgs{
			RealmId:             realm.ID(),
			ClientId:            pulumi.String("frontend"),
			Name:                pulumi.String("Frontend client"),
			Enabled:             pulumi.Bool(true),
			AccessType:          pulumi.String("PUBLIC"),
			StandardFlowEnabled: pulumi.Bool(true),
			ValidRedirectUris: pulumi.StringArray{
				pulumi.String("http://localhost:3000"),
			},
		})

		if err != nil {
			return err
		}

		readAppointmentsScope, err := openid.NewClientScope(ctx, "read:appointments", &openid.ClientScopeArgs{
			RealmId:             realm.ID(),
			Description:         pulumi.String("Read appointments"),
			IncludeInTokenScope: pulumi.Bool(true),
			GuiOrder:            pulumi.Int(1),
		})

		if err != nil {
			return err
		}

		_, err = keycloak.NewGenericRoleMapper(ctx, "read-appointments-role-mapper", &keycloak.GenericRoleMapperArgs{
			RealmId:       realm.ID(),
			RoleId:        memberRole.ID(),
			ClientScopeId: readAppointmentsScope.ID(),
		})

		if err != nil {
			return err
		}

		billingScope, err := openid.NewClientScope(ctx, "write:billing", &openid.ClientScopeArgs{
			RealmId:             realm.ID(),
			Description:         pulumi.String("Write billing"),
			IncludeInTokenScope: pulumi.Bool(true),
			GuiOrder:            pulumi.Int(1),
		})

		if err != nil {
			return err
		}

		_, err = keycloak.NewGenericRoleMapper(ctx, "write-billing-role-mapper", &keycloak.GenericRoleMapperArgs{
			RealmId:       realm.ID(),
			RoleId:        ownerRole.ID(),
			ClientScopeId: billingScope.ID(),
		})

		if err != nil {
			return err
		}

		ctx.Export("realm", realm.Realm)

		return nil
	}
}
