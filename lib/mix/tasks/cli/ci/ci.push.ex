defmodule Mix.Tasks.Ci.Push do
  use Mix.Task

  @shortdoc "Pushes Docker images to GitHub Container Registry"
  def run(_) do
    env = Cli.env()

    IO.puts("pushing images...")
    [env[:app_name], env[:init_name]]
    |> Enum.each(fn image -> push(image, env[:github_owner], env[:git_commit]) end)


  end

  defp push(image, owner, commit) do
    IO.puts("pushing #{image}...")
    Cli.runCommand("docker", ["push", "ghcr.io/#{owner}/#{image}:latest"])
    Cli.runCommand("docker", ["push", "ghcr.io/#{owner}/#{image}:#{commit}"])
  end
end
