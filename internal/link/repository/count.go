package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

func (r *RedisRepository) CountByIP(ctx context.Context, ip string) (int, error) {
	ipKey := ipSetPrefix + ip
	now := time.Now().Unix()

	pipe := r.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, ipKey, "-inf", strconv.FormatInt(now, 10))
	countCommand := pipe.ZCard(ctx, ipKey)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("repository.CountByIp: failed to execute pipeline to count links by ip: %w", err)
	}

	return int(countCommand.Val()), nil
}
