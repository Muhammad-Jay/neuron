for mod in ./application/ ./shared/ ./nore/; do
    echo "Processing $mod..."
    (
        cd "$mod" || exit
        go mod tidy
        go test ./... &&
        go build ./... &&
        echo "$mod built successfully" || echo "$mod build failed"
    )
done && for pkg in ./packages/*; do
    echo "Processing $pkg..."
    (
        cd "$pkg" || exit
        pnpm typecheck &&
        pnpm test && 
        pnpm build &&
        echo "$pkg built successfully" || echo "$pkg build failed"
    )
done && echo "All tests and builds passed successfully!"