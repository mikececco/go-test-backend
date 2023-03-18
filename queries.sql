CREATE FUNCTION spots_bigger_one_domain() RETURN TABLE AS
BEGIN
  RETURN QUERY
    SELECT
      spot_name, --this column includes all spots uniquely
      REGEXP_REPLACE(website, '^(?:https?:\/\/)?(?:www\.)?([^\/]+).*$', '\1') AS domain, --this colum includes domains replaced after regex
      domain_counts.count AS domain_count --this includes domains count (only of domain counted >1)
    FROM spots
    GROUP BY
      spot_name, --group by spot name and domain otherwise we would get back only one row per domain and lose the spot associated to it
      REGEXP_REPLACE(website, '^(?:https?:\/\/)?(?:www\.)?([^\/]+).*$', '\1')
    HAVING
      COUNT(*) > 1; --counts all grouped by and takes just domains with count bigger than 1
END;
