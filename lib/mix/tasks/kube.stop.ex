defmodule Mix.Tasks.Kube.Stop do
  use Mix.Task

  @shortdoc "Stops Kind cluster adn destroys it"
  def run(_) do
    env = Cli.env()

    Cli.runCommand("kind", ["delete", "cluster", "--name", env[:dev_cluster]])
    Cli.runCommand("docker", ["stop", env[:registry_name]])
    Cli.runCommand("docker", ["rm", env[:registry_name]])
  end
end
