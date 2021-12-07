SELECT time, login, content
FROM journal
WHERE time >= @start AND time <= @end
;
