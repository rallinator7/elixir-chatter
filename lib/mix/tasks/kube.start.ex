defmodule Mix.Tasks.Kube.Start do
  use Mix.Task

  @shortdoc "Builds Kind cluster and deploys services into it"
  def run(_) do
    env = Cli.env()

    startRegistry(env[:database_url], env[:secret_key_base] , env[:registry_name])
    createCluster(Path.join([env[:base_dir], "kubernetes"]))
    connectRegistry(env[:registry_name])
    [env[:init_name], env[:app_name]] |> Enum.each(fn image -> pushImage(image) end)
    ["postgres", "chatter"] |> Enum.each(fn chart -> deployService(chart, Path.join([env[:base_dir],"kubernetes", chart])) end)

  end

  defp startRegistry(url, keyBase, name) do
    Cli.runCommand("docker", ["run", "-e", "DATABASE_URL=#{url}", "-e", "SECRET_KEY_BASE=#{keyBase}", "-p",
     "5000:5000", "-d", "--name", name, "registry:2"])
  end

  defp createCluster(path) do
    Cli.runCommand("kind", ["create", "cluster", "--config", "cluster.yaml"], path)
  end

  defp connectRegistry(name) do
    Cli.outputCommand("docker", ["network", "inspect", "kind"])
    |> String.contains?(name)
    |> unless do
       Cli.runCommand("docker", ["network", "connect", "kind", "#{name}"])
    end
  end

  defp pushImage(image) do
    Cli.runCommand("docker", ["tag", "#{image}:latest", "localhost:5000/#{image}:latest"])
    Cli.runCommand("docker", ["push", "localhost:5000/#{image}:latest"])
  end

  defp deployService(chart, path) do
    Cli.runCommand("helm", ["install", chart, "."], path)
  end
end
