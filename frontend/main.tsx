import * as vlens from "vlens";
import * as preact from "preact";
import * as server from "./server";

async function main() {
  vlens.initRoutes([vlens.routeHandler("/", () => import("@app/pages/home/home"))]);
}

main();
