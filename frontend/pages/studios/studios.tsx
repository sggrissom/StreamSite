import * as preact from "preact";
import * as core from "vlens/core";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import "../../styles/global";
import "./studios-styles";

type Data = server.ListMyStudiosResponse;

export async function fetch(route: string, prefix: string) {
  return server.ListMyStudios({});
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  const studios = data?.studios || [];

  return (
    <div>
      <Header />
      <main className="studios-container">
        <div className="studios-content">
          <div className="studios-header">
            <h1 className="studios-title">My Studios</h1>
            <p className="studios-description">
              Studios you are a member of. Create or join studios to start
              streaming.
            </p>
          </div>

          {studios.length === 0 ? (
            <div className="studios-empty">
              <div className="empty-icon">📺</div>
              <h2>No Studios Yet</h2>
              <p>
                You haven't joined any studios yet. Create your first studio to
                get started!
              </p>
              <button
                className="btn btn-primary"
                onClick={() => {
                  // TODO: Open create studio modal
                  alert("Create studio functionality coming soon!");
                }}
              >
                Create Your First Studio
              </button>
            </div>
          ) : (
            <div className="studios-grid">
              {studios.map((studio) => (
                <div key={studio.id} className="studio-card">
                  <div className="studio-card-header">
                    <h3 className="studio-name">{studio.name}</h3>
                    <span className={`studio-role role-${studio.myRole}`}>
                      {studio.myRoleName}
                    </span>
                  </div>

                  {studio.description && (
                    <p className="studio-description">{studio.description}</p>
                  )}

                  <div className="studio-meta">
                    <span className="meta-item">
                      <span className="meta-label">Max Rooms:</span>
                      <span className="meta-value">{studio.maxRooms}</span>
                    </span>
                  </div>

                  <div className="studio-actions">
                    <a
                      href={`/studio/${studio.id}`}
                      className="btn btn-primary btn-sm"
                    >
                      Open Studio
                    </a>
                  </div>
                </div>
              ))}
            </div>
          )}

          {studios.length > 0 && (
            <div className="studios-footer">
              <button
                className="btn btn-primary"
                onClick={() => {
                  // TODO: Open create studio modal
                  alert("Create studio functionality coming soon!");
                }}
              >
                Create New Studio
              </button>
            </div>
          )}
        </div>
      </main>
      <Footer />
    </div>
  );
}
