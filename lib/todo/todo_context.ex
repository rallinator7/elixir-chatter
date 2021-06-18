defmodule Todo.TodoContext do
  @moduledoc """
  The TodoContext context.
  """

  import Ecto.Query, warn: false
  alias Todo.Repo

  alias Todo.TodoContext.TodoSchema

  @doc """
  Returns the list of todos.

  ## Examples

      iex> list_todos()
      [%TodoSchema{}, ...]

  """
  def list_todos do
    Repo.all(TodoSchema)
  end

  @doc """
  Gets a single todo_schema.

  Raises `Ecto.NoResultsError` if the Todo schema does not exist.

  ## Examples

      iex> get_todo_schema!(123)
      %TodoSchema{}

      iex> get_todo_schema!(456)
      ** (Ecto.NoResultsError)

  """
  def get_todo_schema!(id), do: Repo.get!(TodoSchema, id)

  @doc """
  Creates a todo_schema.

  ## Examples

      iex> create_todo_schema(%{field: value})
      {:ok, %TodoSchema{}}

      iex> create_todo_schema(%{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def create_todo_schema(attrs \\ %{}) do
    %TodoSchema{}
    |> TodoSchema.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Updates a todo_schema.

  ## Examples

      iex> update_todo_schema(todo_schema, %{field: new_value})
      {:ok, %TodoSchema{}}

      iex> update_todo_schema(todo_schema, %{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def update_todo_schema(%TodoSchema{} = todo_schema, attrs) do
    todo_schema
    |> TodoSchema.changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Deletes a todo_schema.

  ## Examples

      iex> delete_todo_schema(todo_schema)
      {:ok, %TodoSchema{}}

      iex> delete_todo_schema(todo_schema)
      {:error, %Ecto.Changeset{}}

  """
  def delete_todo_schema(%TodoSchema{} = todo_schema) do
    Repo.delete(todo_schema)
  end

  @doc """
  Returns an `%Ecto.Changeset{}` for tracking todo_schema changes.

  ## Examples

      iex> change_todo_schema(todo_schema)
      %Ecto.Changeset{data: %TodoSchema{}}

  """
  def change_todo_schema(%TodoSchema{} = todo_schema, attrs \\ %{}) do
    TodoSchema.changeset(todo_schema, attrs)
  end
end
