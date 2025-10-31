# Dropbox Connector Setup Guide

---

## Connector capabilities

1. **What resources does the connector sync?**  
   This connector syncs:  
   — Users (Dropbox Team members with full profile information including status and membership type)  
   — Roles (Dropbox Team admin roles for access management)  
   — Groups (Dropbox Team groups with member information)

2. **Can the connector provision any resources? If so, which ones?**  
   The connector can provision:  
   — User accounts (Create new team members via invitation)  
   — User deletion (Remove team members from the organization)  
   — User suspension/activation (Suspend and unsuspend user accounts)  
   — Group membership (Add/remove users from groups)  
   — Role assignments (Assign/revoke admin roles to users)

---

## Connector credentials

1. **What credentials or information are needed to set up the connector?**  
   This connector requires:  
   — Dropbox App Key (OAuth2 Application Key)  
   — Dropbox App Secret (OAuth2 Application Secret)  
   — Dropbox Refresh Token (OAuth2 Refresh Token for API access)

   **Args**:  
   `--app-key` (required)  
   `--app-secret` (required)  
   `--refresh-token` (required)

2. **For each item in the list above:**

   - **How does a user create or look up that credential or info?**

     **Step 1: Access Dropbox App Console**

     - Log in to your Dropbox account at [https://www.dropbox.com](https://www.dropbox.com)
     - Navigate to the **Dropbox App Console**: [https://www.dropbox.com/developers/apps](https://www.dropbox.com/developers/apps)
     - You need to be a **Team Admin** to create apps with team-level permissions

     **Step 2: Create a New App**

     - Click the **"Create app"** button
     - Select **"Scoped access"** (for granular permissions)
     - Choose **"Full Dropbox"** access type
     - Select **"Team"** as the access type (required for managing team members)
     - Enter an **App name** (e.g., "Baton Dropbox Connector")
     - Click **"Create app"**

     **Step 3: Configure App Settings**

     - In the **Settings** tab:
       - Copy the **App key** (use as `--app-key`)
       - Copy the **App secret** (use as `--app-secret`)
     - In the **Permissions** tab:
       - Configure required scopes (see below)
       - Click **"Submit"** to save permission changes

     **Step 4: Generate Refresh Token**

     You can generate a refresh token using the connector's configure mode:

     ```bash
     ./baton-dropbox --app-key=<your-app-key> --app-secret=<your-app-secret> --configure
     ```

     This will:

     - Open a browser window for OAuth authorization
     - Prompt you to log in to Dropbox (if not already logged in)
     - Request permission for the app to access your team
     - Generate and display a **refresh token** (use as `--refresh-token`)
     - Save this token securely - it provides long-term API access

     **Alternative: Manual Token Generation**

     You can also generate tokens manually through the Dropbox OAuth flow:

     1. Navigate to the app's Settings page in the App Console
     2. Under "OAuth 2" section, use the authorization URL
     3. Follow the OAuth flow to obtain an authorization code
     4. Exchange the code for a refresh token using the token endpoint

     **Step 5: Verify Permissions**

     Ensure your app has the required scopes configured (see below).

   - **Does the credential need any specific scopes or permissions?**  
     Yes. The application must have the following Dropbox Team scopes:

     **Available Scopes:**

     - **`team_data.member`**: Read team member information (basic profile, status, roles)
     - **`team_data.governance`**: Read team governance data (roles, permissions)
     - **`team_info.read`**: Read basic team information
     - **`members.write`**: Create and modify team members
     - **`members.delete`**: Remove team members
     - **`groups.write`**: Manage group memberships

     **Required Scopes by Operation:**

     **For Syncing (Read-Only Operations):**

     - `team_data.member` - Read user information and profiles
     - `team_data.governance` - Read role definitions and assignments
     - `team_info.read` - Read team structure and groups

     **For Provisioning (Read-Write Operations):**

     - All sync scopes (above) - Read access for validation
     - `members.write` - Create new team members and suspend/unsuspend accounts
     - `members.delete` - Remove team members from the organization
     - `groups.write` - Add/remove users from groups

   - **Is the list of scopes or permissions different to sync (read) versus provision (read-write)?**  
     Yes, different scopes are required:

     **Syncing Only**: Requires `team_data.member`, `team_data.governance`, `team_info.read`  
     **Provisioning**: Requires all sync scopes PLUS `members.write`, `members.delete`, `groups.write`

     **Recommendation**: For full functionality including provisioning, grant all scopes listed above. The connector will only use provisioning permissions when the `--provisioning` flag is enabled.

   - **What level of access or permissions does the user need in order to create the credentials?**  
     The user must have **Team Admin** role in Dropbox to:
     - Create apps in the Dropbox App Console
     - Configure app scopes and permissions for team-level access
     - Generate OAuth tokens with team management capabilities
     - Access user, role, and group management settings

---

## Additional Notes

### Resource Details

- **Users**: Dropbox Team members including full profile information (email, name, account ID, team member ID, status, membership type). Users can be created (invited), deleted, suspended, and unsuspended.

- **Roles**: Dropbox Team admin roles that define administrative permissions. Supported roles include Team Admin, User Management Admin, Support Admin, and Member Only. Roles can be assigned to users with automatic provisioning support.

- **Groups**: Dropbox Team groups for organizing members and managing shared folder access. Group memberships can be managed programmatically with support for both member and owner access levels.

### User Statuses

This connector handles multiple Dropbox user statuses:

- **Active**: Fully active team member with complete access
- **Invited**: User has been invited but hasn't accepted yet
- **Suspended**: User access has been temporarily suspended
- **Removed**: User has been removed from the team (tracked for audit purposes)

### Supported Provisioning Operations

1. **User Account Management**:

   - **Create Users** (via invitation):

     - **Required fields**: Email Address
     - **Optional fields**: First Name, Last Name (can be set later by user)
     - **Process**: Sends an email invitation to join the team
     - **Note**: User must accept invitation to become active

   - **Delete Users**:

     - Permanently remove users from the team
     - Uses team member ID for identification
     - Automatically handles cleanup of group memberships

   - **Suspend Users**:

     - Temporarily disable user access (account remains in team)
     - User cannot access Dropbox until unsuspended
     - Files and folders remain intact

   - **Unsuspend Users**:
     - Reactivate suspended user accounts
     - Restores full access to files and team resources

2. **Group Membership Management**:

   - Add users to groups (with member or owner permissions)
   - Remove users from groups
   - Supports both regular members and group owners

3. **Role Management**:
   - Assign admin roles to users
   - Revoke admin roles from users
   - Support for multiple admin role types

### Provisioning Field Details

When provisioning users, the connector requires specific fields:

**Required Fields for User Creation**:

- `email` - User's email address (invitation will be sent to this email)

**Note**: Unlike some systems, Dropbox user creation works via invitation. The new user receives an email and must accept the invitation to become an active team member.

### Custom Actions

The connector supports custom actions for advanced user management:

1. **Disable User** (`disable_user`):

   - Suspends a user's access to Dropbox Team
   - **Required argument**: `user_id` (team member ID)
   - **Action type**: Account Disable

2. **Enable User** (`enable_user`):
   - Reactivates a suspended user's access
   - **Required argument**: `user_id` (team member ID)
   - **Action type**: Account Enable

These actions provide fine-grained control over user access beyond standard provisioning operations.