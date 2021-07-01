defmodule Mix.Tasks.Docker.Stop do
  use Mix.Task

  @shortdoc "stops the chatter docker services and database"
  def run(_) do
    env = Cli.env()

    Cli.runCommand("docker", ["stop", env[:app_name]])
    stopCompose(Path.join([env[:base_dir], "docker", "postgres"]))
    Cli.runCommand("docker", ["rm", env[:app_name]])

  end

  defp stopCompose(path) do
    :os.type()
    |> elem(1)
    |> case do
      :linux -> Cli.runCommand("docker-compose", ["down"], path)
      _      -> Cli.runCommand("docker", ["compose", "down"], path)
    end
  end
end
