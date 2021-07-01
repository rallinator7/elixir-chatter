defmodule Mix.Tasks.Ci.UnitTest do
  use Mix.Task

  @shortdoc "Runs unit tests for project"
  def run(_) do
    IO.puts("Building Dependencies...")
    Cli.runCommand("mix", ["deps.get"])

    IO.puts("Running Tests...")
    Cli.runCommand("mix", ["test"])

    IO.puts("All tests passed!")
  end
end
