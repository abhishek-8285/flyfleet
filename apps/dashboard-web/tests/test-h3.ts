import * as h3 from "h3-js";
const polygon: [number, number][] = [
  [8, 68],
  [38, 68],
  [38, 98],
  [8, 98],
  [8, 68]
];
const cells = h3.polygonToCells(polygon, 4);
console.log("Cells count:", cells.length);
if (cells.length > 0) {
  const center = h3.cellToLatLng(cells[0]);
  console.log("First cell lat/lng:", center);
}
const pointCell = h3.latLngToCell(20.5937, 78.9629, 4);
console.log("Point cell:", pointCell);
console.log("Is point in polygon?", cells.includes(pointCell));
