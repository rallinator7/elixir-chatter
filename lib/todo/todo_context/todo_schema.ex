defmodule Todo.TodoContext.TodoSchema do
  use Ecto.Schema
  import Ecto.Changeset

  schema "todos" do
    field :done, :boolean, default: false
    field :title, :string

    timestamps()
  end

  @doc false
  def changeset(todo_schema, attrs) do
    todo_schema
    |> cast(attrs, [:title, :done])
    |> validate_required([:title, :done])
  end
end
