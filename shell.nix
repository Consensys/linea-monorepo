{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.kotlin pkgs.kotlin-language-server pkgs.gradle pkgs.openjdk19 pkgs.gradle
    pkgs.nodejs pkgs.docker-compose
  ];
}
