## mcache

mcache is a simple cache in service memory,and also provide a cache simple cache service. Welcome to use mcache and contribute your code to mcache.

### Design

mcache save cache data in memory , and every item has its expire time, it will be cleaned after expired.

mcache provide these rest api:

- /list : list data cache in table
- /get : get one data cache in table
- /add : put data in cache table
- /delete : delete one data cache
- /update : update one data cache in table
- /register : register a cache in table
