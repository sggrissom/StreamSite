import * as preact from "preact";
import * as server from "../../server";
import { Header, Footer } from "../../layout";
import { StudioHeader } from "./components/StudioHeader";
import { RoomsSection } from "./components/RoomsSection";
import { MembersSection } from "./components/MembersSection";
import "../../styles/global";
import "./studio-styles";

type Data = server.GetStudioDashboardResponse;

export async function fetch(route: string, prefix: string) {
  // Extract studio ID from route (e.g., "/studio/123" -> "123")
  const studioIdStr = route
    .replace(prefix, "")
    .replace(/^\//, "")
    .split("/")[0];
  const studioId = parseInt(studioIdStr, 10);

  return server.GetStudioDashboard({ studioId });
}

export function view(
  route: string,
  prefix: string,
  data: Data,
): preact.ComponentChild {
  // Handle errors or missing data
  if (!data || !data.success) {
    return (
      <div>
        <Header />
        <main className="studio-container">
          <div className="studio-content">
            <div className="error-state">
              <div className="error-icon">⚠️</div>
              <h2>Studio Not Found</h2>
              <p>
                {data?.error ||
                  "The studio you're looking for doesn't exist or you don't have permission to view it."}
              </p>
              <a href="/studios" className="btn btn-primary">
                Back to Studios
              </a>
            </div>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  const studio = data.studio;
  const rooms = data.rooms || [];
  const members = data.members || [];
  const myRole = data.myRole;
  const myRoleName = data.myRoleName;

  // Check if user can manage rooms (Admin or Owner)
  const canManageRooms = myRole >= 2; // Admin or Owner

  return (
    <div>
      <Header />
      <main className="studio-container">
        <div className="studio-content">
          {/* Breadcrumb */}
          <div className="breadcrumb">
            <a href="/studios">Studios</a>
            <span className="breadcrumb-separator">/</span>
            <span className="breadcrumb-current">{studio.name}</span>
          </div>

          {/* Studio Header with metadata and actions */}
          <StudioHeader
            studio={studio}
            myRole={myRole}
            myRoleName={myRoleName}
            rooms={rooms}
            members={members}
            canManageRooms={canManageRooms}
          />

          {/* Rooms Section with all room management */}
          <RoomsSection
            studio={studio}
            rooms={rooms}
            canManageRooms={canManageRooms}
          />

          {/* Members Section with all member management */}
          <MembersSection
            studio={studio}
            members={members}
            canManageRooms={canManageRooms}
          />
        </div>
      </main>
      <Footer />
    </div>
  );
}
