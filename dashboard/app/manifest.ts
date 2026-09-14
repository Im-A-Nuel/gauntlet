import type { MetadataRoute } from "next";

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: "Gauntlet",
    short_name: "Gauntlet",
    description:
      "Inspect mutation testing evidence, surviving mutants, and merge policy.",
    start_url: "/",
    display: "standalone",
    background_color: "#070809",
    theme_color: "#070809",
    icons: [
      {
        src: "/brand/gauntlet-app-icon-192.png",
        sizes: "192x192",
        type: "image/png",
      },
      {
        src: "/brand/gauntlet-app-icon-512.png",
        sizes: "512x512",
        type: "image/png",
      },
    ],
  };
}
