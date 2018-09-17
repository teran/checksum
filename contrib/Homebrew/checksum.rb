class Checksum < Formula
  desc "SHA256 file verification for consistency check purposes"
  homepage "https://github.com/teran/checksum"
  version "0.6"
  url "https://github.com/teran/checksum/archive/v#{version}.tar.gz"
  sha256 "0f77ac41bf42c9566f4a3e907de74898c2bfeb80449cefd4ef25c1252a875467"

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
