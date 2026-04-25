for dir in service/*/; do
  if [ -f "$dir/internal/conf/conf.proto" ]; then
    echo "Generating config for $dir..."
    (cd "$dir" && make config 2>/dev/null || \
     protoc --proto_path=. --proto_path=../../third_party \
       --go_out=. --go_opt=paths=source_relative \
       internal/conf/conf.proto)
  fi
done