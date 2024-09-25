import { kysely } from "../src/database";
import { populateProductTable } from "../src/job";

console.log("Populating product table...");
await populateProductTable();
console.log("Done!");

await kysely.destroy();
