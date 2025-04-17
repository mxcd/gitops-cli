class Gitops < Formula
  desc "CLI tool for performing gitops tasks"
  homepage "https://github.com/mxcd/gitops-cli"
  url "https://github.com/mxcd/gitops-cli/archive/${RELEASE_VERSION}.tar.gz"
  sha256 "${RELEASE_SHA256}"
  license "MIT"
  depends_on "go" => :build

  def install
    ldflags = "-s -w -X main.version=#{version}"
    cd "cmd/gitops" do
      system "go", "build", *std_go_args(ldflags: ldflags)
    end
  end

  test do
    shell_output("#{bin}/gitops", "--version")
  end
end