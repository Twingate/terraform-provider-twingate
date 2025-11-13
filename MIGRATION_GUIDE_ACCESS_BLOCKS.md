# Migration Guide: Access Blocks Changed from Set to List

## Overview

In version **[NEXT_VERSION]**, we are changing the `access_service` and `access_group` blocks from **Set** types to **List** types in the `twingate_resource` resource.

## Why This Change?

We discovered a compatibility issue between the Terraform Plugin Framework's Set types and the Pulumi Terraform Bridge. When using Set types for nested blocks:

- **Terraform Direct**: Works correctly ✅
- **Pulumi**: Only returns the first element during Read operations ❌

This bug manifests during:
- Resource refresh operations
- Subsequent runs after initial creation
- State exports

By changing from Set to List, we ensure:
- **Full compatibility** with both Terraform and Pulumi
- **All service accounts and groups** are correctly returned
- **Consistent behavior** across all infrastructure-as-code tools

## What's Changing?

### Before (Set-based)
```hcl
resource "twingate_resource" "example" {
  name              = "Example Resource"
  address           = "example.com"
  remote_network_id = twingate_remote_network.example.id

  # These were Sets - order not guaranteed
  access_service {
    service_account_id = twingate_service_account.account1.id
  }

  access_service {
    service_account_id = twingate_service_account.account2.id
  }

  access_group {
    group_id = twingate_group.group1.id
  }

  access_group {
    group_id = twingate_group.group2.id
  }
}
```

### After (List-based)
```hcl
resource "twingate_resource" "example" {
  name              = "Example Resource"
  address           = "example.com"
  remote_network_id = twingate_remote_network.example.id

  # These are now Lists - order is preserved
  access_service {
    service_account_id = twingate_service_account.account1.id
  }

  access_service {
    service_account_id = twingate_service_account.account2.id
  }

  access_group {
    group_id = twingate_group.group1.id
  }

  access_group {
    group_id = twingate_group.group2.id
  }
}
```

## Impact on Your Infrastructure

### HCL Syntax
**No changes required** - the HCL syntax remains identical. You can keep using your existing Terraform configurations without modification.

### State Files
The Terraform provider includes automatic state migration. When you upgrade to the new version:

1. **First `terraform plan`**: Terraform will detect the state format change and automatically migrate your state
2. **No manual intervention** required
3. **No resource recreation** - this is a state-only change

### Behavioral Changes

#### Order Preservation
- **Before**: The order of `access_service` and `access_group` blocks was not guaranteed (Set semantics)
- **After**: The order is now preserved (List semantics)

**Impact**: If your configurations rely on a specific order, that order will now be maintained consistently.

#### Duplicate Detection
- **Before**: Sets automatically prevented duplicates
- **After**: Lists allow duplicates (though Twingate API will still reject them)

**Impact**: Minimal - the Twingate API validation remains unchanged. If you accidentally specify the same group or service account twice, you'll get an API error (same as before).

## Migration Steps

### For Terraform Users

1. **Upgrade the Provider**:
   ```hcl
   terraform {
     required_providers {
       twingate = {
         source  = "Twingate/twingate"
         version = "~> [NEXT_VERSION]"
       }
     }
   }
   ```

2. **Run Terraform Init**:
   ```bash
   terraform init -upgrade
   ```

3. **Review the Plan**:
   ```bash
   terraform plan
   ```

   You should see a message about state migration, but **no resource changes**.

4. **Apply the Changes**:
   ```bash
   terraform apply
   ```

   This will update your state file to the new format.

### For Pulumi Users

This change **fixes the bug** where only the first service account or group was being returned!

1. **Upgrade the Pulumi Provider**:
   ```bash
   # For JavaScript/TypeScript
   npm install @twingate/pulumi-twingate@latest

   # For Python
   pip install --upgrade pulumi-twingate

   # For Go
   go get github.com/Twingate/pulumi-twingate/sdk/v3@latest
   ```

2. **Run Pulumi Refresh**:
   ```bash
   pulumi refresh
   ```

   You should now see **all** service accounts and groups in your state, not just the first one!

3. **Verify the Fix**:
   ```bash
   pulumi stack export | jq '.deployment.resources[] | select(.type == "twingate:index/twingateResource:TwingateResource") | {inputs: .inputs.accessServices, outputs: .outputs.accessServices}'
   ```

   Both `inputs` and `outputs` should now show all items.

## Rollback Procedure

If you need to rollback to the previous version:

1. **Pin to the Old Version**:
   ```hcl
   terraform {
     required_providers {
       twingate = {
         source  = "Twingate/twingate"
         version = "= [PREVIOUS_VERSION]"
       }
     }
   }
   ```

2. **Re-initialize**:
   ```bash
   terraform init -upgrade
   ```

3. **Restore State** (if needed):
   If you have state backup, you can restore it:
   ```bash
   terraform state pull > current-state.json  # Backup current state first
   # Restore from backup if needed
   ```

## Frequently Asked Questions

### Q: Do I need to modify my existing Terraform configurations?
**A**: No, the HCL syntax remains unchanged. Only the internal state representation is different.

### Q: Will my resources be recreated?
**A**: No, this is a state-only change. No resources will be destroyed or recreated.

### Q: What if I have a very large state file?
**A**: The migration is automatic and should complete quickly. State file size doesn't significantly impact migration time.

### Q: Will this affect my existing resources in the Twingate console?
**A**: No, this change only affects how Terraform and Pulumi manage state. Your actual Twingate resources remain unchanged.

### Q: I'm using Pulumi and only see one service account - is this related?
**A**: Yes! This change fixes that exact issue. After upgrading, all service accounts and groups will be correctly reflected in your Pulumi state.

### Q: Can I have both the old and new provider versions in different workspaces?
**A**: Yes, but be careful with shared state backends. Each workspace maintains its own state, so different workspaces can use different provider versions safely.

### Q: What happens if I specify the same service account twice?
**A**: The Twingate API will reject the duplicate, and you'll receive an error message (same behavior as before).

## Need Help?

If you encounter any issues during migration:

1. **Check the logs**: Run Terraform/Pulumi with debug logging enabled
   ```bash
   TF_LOG=DEBUG terraform apply
   # or
   pulumi up --logtostderr -v=9
   ```

2. **Open an issue**: https://github.com/Twingate/terraform-provider-twingate/issues

3. **Contact support**: support@twingate.com

## Technical Details

For those interested in the technical implementation:

- Changed `schema.SetNestedBlock` to `schema.ListNestedBlock`
- Updated field types from `types.Set` to `types.List`
- Modified conversion functions to use `makeObjectsList` instead of `makeObjectsSet`
- Updated validators from `setvalidator` to `listvalidator`
- All helper functions now work with List semantics
- State migration code automatically converts old Set-based state to new List-based state

This change improves compatibility with the Pulumi Terraform Bridge while maintaining full backward compatibility for Terraform users.
