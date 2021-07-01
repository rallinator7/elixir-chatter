defmodule Cli do

  # returns the set environment variables for building and running the project
  def env() do
    %{
      :postgres_network => "phoenix-postgres",
      :postgres_volume  => "phoenix_postgres",
      :github_owner     => "rallinator7",
      :app_name         => "chatter",
      :git_commit       => gitCommit(),
      :base_dir         => File.cwd!,
      :init_name        => "phoenix-init",
      :dtabase_url      => "ecto://phoenix:phoenix@db:5432/phoenix",
      :secret_key_base  => "JhhLO9oACpINDgzWo9xBWw+qKCrh7C6tzUhBo4rMGCbB51ssgPzZpkL812d12fL1",
      :dev_cluster      => "dev-cluster",
      :registry_name    => "kind-registry",
      :registry_port    => "registry_port"
    }
  end

  def runCommand(command, args, out \\ false) do
    cmdReturn = System.cmd(command, args)
    exitStatus = elem(cmdReturn, 1)

    if exitStatus > 0 do
      exit({:shutdown, exitStatus})
    end

    if out do
      cmdReturn |> elem(0) |> String.trim() |> IO.puts()
    end
  end

  defp gitCommit() do
    {hash, _} = System.cmd("git", ["rev-parse", "--short", "HEAD"])

    String.trim(hash)
  end

end
