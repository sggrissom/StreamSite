import * as vlens from "vlens";
import * as preact from "preact";
import * as server from "./server";

async function main() {
  vlens.initRoutes([
    vlens.routeHandler("/login", () => import("@app/pages/auth/login")),
    vlens.routeHandler(
      "/create-account",
      () => import("@app/pages/auth/create-account"),
    ),
    vlens.routeHandler(
      "/dashboard",
      () => import("@app/pages/dashboard/dashboard"),
    ),
    vlens.routeHandler("/studios", () => import("@app/pages/studios/studios")),
    vlens.routeHandler(
      "/stream-admin",
      () => import("@app/pages/stream-admin/stream-admin"),
    ),
    vlens.routeHandler(
      "/site-admin",
      () => import("@app/pages/site-admin/site-admin"),
    ),
    vlens.routeHandler(
      "/settings",
      () => import("@app/pages/settings/settings"),
    ),
    vlens.routeHandler("/stream", () => import("@app/pages/stream/stream")),
    vlens.routeHandler("/", () => import("@app/pages/home/home")),
  ]);
}

main();
