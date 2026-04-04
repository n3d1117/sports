class Sports < Formula
  desc "CLI for sports data"
  homepage "https://github.com/n3d1117/sports"
  url "https://github.com/n3d1117/sports/archive/refs/tags/v1.0.1.tar.gz"
  sha256 "2703d3edca8b1ee5bc66ec9dfa6d7c2095d580ceeec8af2cec82f6fb9aaea9b8"
  license "MIT"
  head "https://github.com/n3d1117/sports.git", branch: "main"

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
