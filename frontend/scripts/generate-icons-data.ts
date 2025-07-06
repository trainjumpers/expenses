import * as fs from "fs";
import * as path from "path";

// Paths
const iconsDir = "./lucide/icons";
const outputFile = path.resolve(__dirname, "../components/ui/icons-data.ts");

// Read all JSON files in the icons directory
const files = fs.readdirSync(iconsDir).filter((f) => f.endsWith(".json"));

const iconsData: Array<{ name: string; categories: string[]; tags: string[] }> =
  [];

for (const file of files) {
  const filePath = path.join(iconsDir, file);
  const content = fs.readFileSync(filePath, "utf-8");
  try {
    const json = JSON.parse(content);
    if (Array.isArray(json.categories) && Array.isArray(json.tags)) {
      iconsData.push({
        name: path.basename(file, ".json"),
        categories: json.categories,
        tags: json.tags,
      });
    } else {
      console.warn(`Skipping ${file}: missing categories or tags array`);
    }
  } catch (e) {
    console.error(`Error parsing ${file}:`, e);
  }
}

// Optionally, preserve the IconCategory enum if it exists
let enumContent = "";
if (fs.existsSync(outputFile)) {
  const existing = fs.readFileSync(outputFile, "utf-8");
  const match = existing.match(/export enum IconCategory[\s\S]+?\n}\n/);
  if (match) {
    enumContent = match[0] + "\n";
  }
}

const output =
  enumContent +
  `export const iconsData: Array<{ name: string; categories: string[]; tags: string[] }> = ${JSON.stringify(iconsData, null, 2)};\n`;

fs.writeFileSync(outputFile, output, "utf-8");

console.log(`Wrote ${iconsData.length} icons to ${outputFile}`);
