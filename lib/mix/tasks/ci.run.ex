defmodule Mix.Tasks.Ci.Run do
  use Mix.Task

  @shortdoc "Runs unit tests, build docker images, and pushes to ghcr"
  def run(_) do
    env = Cli.env()

    test()
    build(env)
    push(env)


  end

  defp test() do
    IO.puts("Running Tests...")
    Cli.runCommand("mix", ["test"])

    IO.puts("All tests passed!")
  end

  defp build(env) do
    IO.puts("Building #{env[:app_name]} service...")
    Cli.runCommand("docker", ["build", "--build-arg", "DATABASE_URL=#{env[:database_url]}", "--build-arg", "SECRET_KEY_BASE=#{env[:secret_key_base]}",
    "-t", "ghcr.io/#{env[:github_owner]}/#{env[:app_name]}:#{env[:git_commit]}", "-t", "ghcr.io/#{env[:github_owner]}/#{env[:app_name]}:latest", "-f", "./docker/server/Dockerfile", "."])

    IO.puts("Building #{env[:init_name]} service...")
    Cli.runCommand("docker", ["build", "--build-arg", "DATABASE_URL=#{env[:database_url]}", "--build-arg", "SECRET_KEY_BASE=#{env[:secret_key_base]}",
    "-t", "ghcr.io/#{env[:github_owner]}/#{env[:init_name]}:#{env[:git_commit]}", "-t", "ghcr.io/#{env[:github_owner]}/#{env[:init_name]}:latest", "-f", "./docker/init/Dockerfile", "."])
  end

  defp push(env) do
    IO.puts("pushing images...")
    [env[:app_name], env[:init_name]]
    |> Enum.each(fn image -> executePush(image, env[:github_owner], env[:git_commit]) end)
  end

  defp executePush(image, owner, commit) do
    IO.puts("pushing #{image}...")
    Cli.runCommand("docker", ["push", "ghcr.io/#{owner}/#{image}:latest"])
    Cli.runCommand("docker", ["push", "ghcr.io/#{owner}/#{image}:#{commit}"])
  end
end
