# Data Model: M28 Artifact Runtime Foundation

## Artifact

- `id`: stable local artifact id
- `thread_id`: owning thread
- `run_id`: source run
- `title`: safe display title
- `artifact_type`: `text`
- `content`: bounded UTF-8 text stored by productdata
- `content_bytes`: byte length
- `excerpt`: bounded safe text excerpt
- `created_at`: service timestamp
- `updated_at`: service timestamp

## Artifact Tool Arguments

`artifact.create_text`:

- `title`: required safe title
- `content`: required bounded UTF-8 text
- `max_bytes`: optional content byte cap

`artifact.read`:

- `artifact_id`: required existing artifact id
- `max_bytes`: optional excerpt byte cap

`artifact.list`:

- `limit`: optional bounded result count

## Artifact Result Summary

- `tool`
- `scope = artifact`
- `operation`
- `artifact_id`
- `title`
- `artifact_type`
- `content_bytes`
- `text_excerpt`
- `truncated`
- `redaction_applied`
