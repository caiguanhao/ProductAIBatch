ProductAIBatch
--------------

Retrieve ProductAI results and put them into PostgreSQL.

```
echo "https://upload.wikimedia.org/wikipedia/commons/thumb/d/d9/Kubus_sofa.jpg/800px-Kubus_sofa.jpg" > url_file

ProductAIBatch --service-id 00000000 --access-key-id 00000000000000000000000000000000 url_file >sql 2>log
```

The `sql` file example:

```
https://upload.wikimedia.org/wikipedia/commons/thumb/d/d9/Kubus_sofa.jpg/800px-Kubus_sofa.jpg    null    {"coordinates":[[0.11375,0.07111111111111111,0.815,0.9022222222222223]],"results":[{"image_url":"http://example.com/match.jpg","id":"10000001"}]}
https://upload.wikimedia.org/wikipedia/commons/thumb/d/d9/Kubus_sofa.jpg/800px-Kubus_sofa.jpg    ["0.11375","0.07111111","0.815","0.9022222"]    {"coordinates":[[0.11375,0.07111111,0.815,0.9022222]],"results":[{"image_url":"http://example.com/match.jpg","id":"10000001"}]}
```

Create database `ai` and table:

```
DROP TABLE IF EXISTS airesults;
CREATE TABLE airesults (
  url character varying NOT NULL,
  coords json,
  result json NOT NULL
);
```

Import to database:

```
awk 'BEGIN{print "COPY airesults (url, coords, result) FROM stdin;"} {print} END{print "\\."}' sql | psql ai
```

To get all images from your database:

```
\copy (SELECT (regexp_matches(body, '<img[^>]+src="([^"]+)"[^>]+>', 'g'))[1] FROM "posts" WHERE "posts"."deleted_at" IS NULL AND "posts"."status" = 'published') to '/tmp/list'
```

You can sort and remove duplicated lines.  If your file is large, split it:

```
split -l 10000 list
```

Run multiple instances (in different panes in tmux):

```
ProductAIBatch lists/xaa >> done/xaa
ProductAIBatch lists/xab >> done/xab
ProductAIBatch lists/xac >> done/xac
...
```
