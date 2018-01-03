class Checksum < Formula
  desc "SHA256 file verification for consistency check purposes"
  homepage "https://github.com/teran/checksum"
  version "0.4"
  url "https://github.com/teran/checksum/archive/v#{version}.tar.gz"
  sha256 "085cc8a0feba7eb3341a7476692c5edf4136a2463f93f305be620a5687705d2a"

  depends_on "go" => :build
  depends_on "make" => :build

  def install
    ENV["GOPATH"] = buildpath
    ENV["REVISION"] = version
    mkdir_p "src/github.com/teran"
    ln_s buildpath, "src/github.com/teran/checksum"
    system "make", "build-macos-amd64"
    system "mv", "bin/checksum-darwin-amd64", "bin/checksum"
    bin.install "bin/checksum"
  end

  test do
    system "make", "test"
  end
end
