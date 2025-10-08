import * as vlens from "vlens";
import * as preact from "preact";
import * as server from "./server";

async function main() {
  vlens.initRoutes([
    vlens.routeHandler("/login", () => import("@app/pages/auth/login")),
    vlens.routeHandler("/create-account", () => import("@app/pages/auth/create-account")),
    vlens.routeHandler("/dashboard", () => import("@app/pages/dashboard/dashboard")),
    vlens.routeHandler("/stream", () => import("@app/pages/stream/stream")),
    vlens.routeHandler("/", () => import("@app/pages/home/home")),
  ]);
}

main();
