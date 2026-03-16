#!/usr/bin/env bash
set -e

cd "$(dirname "$0")"

GO=${GO:-$(which go 2>/dev/null || echo "$HOME/go/go1.24.2/bin/go")}

echo "→ Stopping running instance..."
pkill -x chekout 2>/dev/null || true
sleep 0.3

echo "→ Building..."
"$GO" build -o chekout .

echo "→ Updating bundle..."
mkdir -p chekout.app/Contents/MacOS
cp chekout chekout.app/Contents/MacOS/

# Only copy Info.plist and register if the bundle is new
if [ ! -f chekout.app/Contents/Info.plist ]; then
  cp Info.plist chekout.app/Contents/
  echo "→ Registering URL scheme..."
  /System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/LaunchServices.framework/Versions/A/Support/lsregister -f "$(pwd)/chekout.app"
fi

echo "→ Launching..."
open chekout.app
echo "✓ Done"
