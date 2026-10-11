class Sports < Formula
  desc "CLI for sports data"
  homepage "https://github.com/n3d1117/sports"
  url "https://github.com/n3d1117/sports/archive/refs/tags/v1.0.3.tar.gz"
  sha256 "6055f5a93021e4fa2bed7e64397c61707dad6ca34be09b9f0e2b066e2cd1c889"
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
