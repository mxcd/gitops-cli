class Gitops < Formula
  desc "CLI tool for performing gitops tasks"
  homepage "https://github.com/mxcd/gitops-cli"
  url "https://github.com/mxcd/gitops-cli/archive/${RELEASE_VERSION}.tar.gz"
  sha256 "${RELEASE_SHA256}"
  license "MIT"
  depends_on "go" => :build

  def install
    # ENV["GOPATH"] = buildpath
    path = buildpath/"cmd/gitops"
    cd path do
      system "go", "build", "-ldflags=\"-s -w -X 'main.version=${RELEASE_VERSION}'\"" "-o", "#{bin}/gitops"
    end
  end

  test do
    shell_output("#{bin}/gitops", "-h")
  end
end