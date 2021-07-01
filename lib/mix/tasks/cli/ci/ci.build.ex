defmodule Mix.Tasks.Ci.Build do
  use Mix.Task

  @shortdoc "Builds Docker images"
  def run(_) do
    env = Cli.env()

    IO.puts("Building #{env[:app_name]} service...")
    Cli.runCommand("docker", ["build", "--build-arg", "DATABASE_URL=#{env[:database_url]}", "--build-arg", "SECRET_KEY_BASE=#{env[:secret_key_base]}",
    "-t", "ghcr.io/#{env[:github_owner]}/#{env[:app_name]}:#{env[:git_commit]}", "-t", "ghcr.io/#{env[:github_owner]}/#{env[:app_name]}:latest", "-f", "./docker/server/Dockerfile", "."])

    IO.puts("Building #{env[:init_name]} service...")
    Cli.runCommand("docker", ["build", "--build-arg", "DATABASE_URL=#{env[:database_url]}", "--build-arg", "SECRET_KEY_BASE=#{env[:secret_key_base]}",
    "-t", "ghcr.io/#{env[:github_owner]}/#{env[:init_name]}:#{env[:git_commit]}", "-t", "ghcr.io/#{env[:github_owner]}/#{env[:init_name]}:latest", "-f", "./docker/init/Dockerfile", "."])
  end
end
