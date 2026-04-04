class Sports < Formula
  desc "CLI for sports data"
  homepage "https://github.com/n3d1117/sports"
  url "https://github.com/n3d1117/sports/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "1ab90a8697e8a65ae6943176511058874b60ad7e4dc3d2211f7bef48b012d96a"
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
