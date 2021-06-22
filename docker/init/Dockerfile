FROM elixir:1.12-alpine

COPY . .

ARG DATABASE_URL

ARG SECRET_KEY_BASE

ENV DATABASE_URL=${DATABASE_URL}

ENV SECRET_KEY_BASE=${SECRET_KEY_BASE}

ENV MIX_ENV=prod

RUN mix local.hex --force && mix deps.get && mix local.rebar --force

CMD [ "mix", "ecto.setup" ]