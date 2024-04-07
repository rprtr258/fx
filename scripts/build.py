import asyncio
from subprocess import run

import requests

async def main():
  latest = requests.get("https://api.github.com/repos/rprtr258/fx/releases/latest", timeout=10).json()["tag_name"]

  run("go mod download")
  async for goos in ["linux", "darwin", "windows"]:
    async for goarch in ["amd64", "arm64"]:
      name = f"fx_{goos}_{goarch}" + (".exe" if goos == "windows" else "")
      run(f"GOOS={goos} GOARCH={goarch} go build -o {name}")
      run(f"gh release upload {latest} {name}")
      run(f"rm {name}")

if __name__ == "__main__":
    asyncio.run(main())