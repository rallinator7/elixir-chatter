defmodule Mix.Tasks.Docker.Start do
  use Mix.Task

  @shortdoc "Builds Docker images, prepare docker-compose, and finally runs the server"
  def run(_) do
    env = Cli.env()

    getDependencies()
    setComposeRequirements(env[:postgres_network], env[:postgres_volume])
    buildImage(env[:init_name], env[:git_commit], env[:database_url], env[:secret_key_base], Path.join([env[:base_dir], "docker", "init", "Dockerfile"]))
    buildImage(env[:app_name], env[:git_commit], env[:database_url], env[:secret_key_base], Path.join([env[:base_dir], "docker", "server", "Dockerfile"]))
    startCompose(Path.join([env[:base_dir], "docker", "postgres"]))
    runInit(env[:database_url], env[:secret_key_base], env[:postgres_network], env[:init_name])
    runServer(env[:database_url], env[:secret_key_base], env[:postgres_network], env[:app_name])

  end

  defp getDependencies() do
    Cli.runCommand("mix", ["deps.get"])
  end

  defp setComposeRequirements(network, volume) do
    Cli.outputCommand("docker", ["network", "ls"])
    |> String.contains?(network)
    |> unless do
       Cli.runCommand("docker", ["network", "create", network])
    end

    Cli.outputCommand("docker", ["volume", "ls"])
    |> String.contains?(volume)
    |> unless do
       Cli.runCommand("docker", ["volume", "create", volume])
    end
  end

  defp buildImage(image, commit, url, keyBase, path) do
    IO.puts("Building #{image}...")

    Cli.runCommand("docker", ["build", "--build-arg", "DATABASE_URL=#{url}", "--build-arg", "SECRET_KEY_BASE=#{keyBase}",
    "-t", "#{image}:#{commit}", "-t", "#{image}:latest", "-f", path, "."])
  end

  defp startCompose(path) do
    :os.type()
    |> elem(1)
    |> case do
      :linux -> Cli.runCommand("docker-compose", ["up", "-d"], path)
      _      -> Cli.runCommand("docker", ["compose", "up", "-d"], path)
    end
  end

  defp runInit(url, keyBase, network, name) do
    Cli.runCommand("docker", ["run", "-e", "DATABASE_URL=#{url}", "-e", "SECRET_KEY_BASE=#{keyBase}",
    "--network", network, "#{name}:latest"])
  end

  defp runServer(url, keyBase, network, name) do
    Cli.runCommand("docker", ["run", "-e", "DATABASE_URL=#{url}", "-e", "SECRET_KEY_BASE=#{keyBase}",
    "--network", network, "-p", "4000:4000", "-d", "--name", name, "#{name}:latest"])
  end
end
