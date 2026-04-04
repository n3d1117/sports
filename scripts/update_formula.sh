#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <tag> <owner/repo>" >&2
  exit 1
fi

tag="$1"
repo="$2"
formula_path="Formula/sports.rb"
source_url="https://github.com/${repo}/archive/refs/tags/${tag}.tar.gz"
tmp_tarball="$(mktemp)"

cleanup() {
  rm -f "$tmp_tarball"
}
trap cleanup EXIT

curl -fsSL "$source_url" -o "$tmp_tarball"

sha="$(shasum -a 256 "$tmp_tarball" | awk '{print $1}')"

cat >"$formula_path" <<EOF
class Sports < Formula
  desc "CLI for sports data"
  homepage "https://github.com/${repo}"
  url "${source_url}"
  sha256 "${sha}"
  license "MIT"
  head "https://github.com/${repo}.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X sports/internal/buildinfo.Version=#{version}"
    system "go", "build", *std_go_args(output: bin/"sports", ldflags: ldflags), "./cmd/sports"
    system "go", "build", *std_go_args(output: bin/"sports-mcp", ldflags: ldflags), "./cmd/sports-mcp"
  end

  test do
    assert_match "Sports data CLI", shell_output("#{bin}/sports --help")
    assert_match "Usage: sports-mcp", shell_output("#{bin}/sports-mcp --help")
  end
end
EOF
