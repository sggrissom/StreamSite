import * as preact from "preact";
import * as vlens from "vlens";
import * as rpc from "vlens/rpc";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./site-admin-styles";

type Data = {
  auth: server.AuthResponse | null;
  users: server.UserListInfo[];
};

export async function fetch(route: string, prefix: string) {
  // Check if user is authenticated and has site admin role
  let [authResp, authErr] = await server.GetAuthContext({});

  let users: server.UserListInfo[] = [];
  if (authResp && authResp.isSiteAdmin) {
    let [usersResp, usersErr] = await server.ListUsers({});
    if (!usersErr && usersResp) {
      users = usersResp.users || [];
    }
  }

  return rpc.ok<Data>({
    auth: authResp || null,
    users: users,
  });
}

const state = vlens.declareHook(() => {
  return {
    selectedUserId: 0,
    selectedRole: 0,
    updating: false,
    error: "",
  };
});

async function handleRoleChange(userId: number, newRole: number) {
  const s = state();
  s.updating = true;
  s.error = "";
  vlens.scheduleRedraw();

  const [resp, err] = await server.SetUserRole({
    userId: userId,
    role: newRole,
  });

  s.updating = false;

  if (err || !resp || !resp.success) {
    s.error = resp?.error || err || "Failed to update role";
    vlens.scheduleRedraw();
    return;
  }

  // Reload the page to refresh user list
  window.location.reload();
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  const s = state();

  // Redirect to login if not authenticated
  if (!data.auth || data.auth.id === 0) {
    core.setRoute("/login");
    return <div></div>;
  }

  // Redirect to regular dashboard if not site admin
  if (!data.auth.isSiteAdmin) {
    core.setRoute("/dashboard");
    return <div></div>;
  }

  return (
    <div>
      <Header />
      <main className="site-admin-container">
        <div className="site-admin-content">
          <h1 className="site-admin-title">Site Admin Dashboard</h1>
          <p className="site-admin-description">
            Welcome, {data.auth.name}! You have full Site Admin access.
          </p>

          {s.error && (
            <div className="error-message">
              {s.error}
            </div>
          )}

          <div className="admin-sections">
            <div className="admin-section full-width">
              <h2>User Management</h2>
              <p>Manage user roles and permissions.</p>

              <div className="users-table">
                <table>
                  <thead>
                    <tr>
                      <th>ID</th>
                      <th>Name</th>
                      <th>Email</th>
                      <th>Current Role</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.users.map((user) => (
                      <tr key={user.id}>
                        <td>{user.id}</td>
                        <td>{user.name}</td>
                        <td>{user.email}</td>
                        <td>
                          <span className={`role-badge role-${user.role}`}>
                            {user.roleName}
                          </span>
                        </td>
                        <td>
                          {user.id !== data.auth!.id && (
                            <select
                              disabled={s.updating}
                              onChange={(e) => {
                                const target = e.target as HTMLSelectElement;
                                handleRoleChange(user.id, parseInt(target.value));
                              }}
                              value={user.role}
                            >
                              <option value={0}>User</option>
                              <option value={1}>Stream Admin</option>
                              <option value={2}>Site Admin</option>
                            </select>
                          )}
                          {user.id === data.auth!.id && (
                            <span className="text-muted">(You)</span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            <div className="admin-section">
              <h2>System Settings</h2>
              <p>Configure site-wide settings and preferences.</p>
              <div className="section-actions">
                <button className="btn btn-secondary">Manage Settings</button>
              </div>
            </div>

            <div className="admin-section">
              <h2>Analytics</h2>
              <p>View site usage statistics and user activity.</p>
              <div className="section-actions">
                <button className="btn btn-secondary">View Analytics</button>
              </div>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
