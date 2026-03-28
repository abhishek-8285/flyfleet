func (r *AnalyticsRepositoryImpl) ListDriverBehaviorAnalytics(ctx context.Context, orgID uuid.UUID, periodStart, periodEnd *time.Time, cursor string, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	// Count query
	countQuery := "SELECT COUNT(*) FROM driver_behavior_analytics WHERE organization_id = $1"
	countArgs := []interface{}{orgID}
	
	// Filter conditions
	whereConditions := []string{"organization_id = $1"}
	argIndex := 2
	
	if periodStart != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("analysis_period_start >= $%d", argIndex))
		countArgs = append(countArgs, *periodStart)
		argIndex++
	}
	
	if periodEnd != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("analysis_period_end <= $%d", argIndex))
		countArgs = append(countArgs, *periodEnd)
		argIndex++
	}
	
	finalCountQuery := countQuery
	if len(whereConditions) > 1 {
		finalCountQuery += " AND " + strings.Join(whereConditions[1:], " AND ")
	}
	
	var totalCount int
	err := r.db.QueryRow(ctx, finalCountQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count driver behavior analytics: %w", err)
	}
	
	// Main query
	mainQuery := `
		SELECT id, driver_id, organization_id, analysis_period_start, analysis_period_end, 
			total_driving_hours, aggressive_acceleration_count, hard_braking_count, sharp_turn_count, 
			over_speeding_count, idling_time, average_speed, max_speed, fuel_efficiency, 
			safety_score, efficiency_score, comfort_score, ranking, improvement_suggestions, 
			behavior_trend, generated_at, created_at, updated_at
		FROM driver_behavior_analytics 
		WHERE organization_id = $1
	`
	
	args := []interface{}{orgID}
	argIndex = 2
	
	whereConditions = []string{} // Reset for main query
	if periodStart != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("analysis_period_start >= $%d", argIndex))
		args = append(args, *periodStart)
		argIndex++
	}
	
	if periodEnd != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("analysis_period_end <= $%d", argIndex))
		args = append(args, *periodEnd)
		argIndex++
	}
	
	if cursor != "" {
		cursorTime, cursorID, err := utils.DecodeCursor(cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid cursor: %w", err)
		}
		whereConditions = append(whereConditions, fmt.Sprintf("(analysis_period_start, id) < ($%d, $%d)", argIndex, argIndex+1))
		args = append(args, cursorTime, cursorID)
		argIndex += 2
	}

	finalMainQuery := mainQuery
	if len(whereConditions) > 0 {
		finalMainQuery += " AND " + strings.Join(whereConditions, " AND ")
	}
	
	if cursor != "" {
		finalMainQuery += fmt.Sprintf(" ORDER BY analysis_period_start DESC, id DESC LIMIT $%d", argIndex)
		args = append(args, limit)
	} else {
		finalMainQuery += fmt.Sprintf(" ORDER BY analysis_period_start DESC, id DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, limit, offset)
	}
	
	rows, err := r.executeWithObservability(ctx, finalMainQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query driver behavior analytics: %w", err)
	}
	defer rows.Close()
	
	var analyticsList []*models.DriverBehaviorAnalytics
	for rows.Next() {
		var analytics models.DriverBehaviorAnalytics
		var improvementSuggestionsJSON []byte
		
		err := rows.Scan(
			&analytics.ID, &analytics.DriverID, &analytics.OrganizationID, &analytics.AnalysisPeriodStart,
			&analytics.AnalysisPeriodEnd, &analytics.TotalDrivingHours, &analytics.AggressiveAccelerationCount,
			&analytics.HardBrakingCount, &analytics.SharpTurnCount, &analytics.OverSpeedingCount,
			&analytics.IdlingTime, &analytics.AverageSpeed, &analytics.MaxSpeed, &analytics.FuelEfficiency,
			&analytics.SafetyScore, &analytics.EfficiencyScore, &analytics.ComfortScore, &analytics.Ranking,
			&improvementSuggestionsJSON, &analytics.BehaviorTrend, &analytics.GeneratedAt,
			&analytics.CreatedAt, &analytics.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan driver behavior analytics row: %w", err)
		}
		
		err = json.Unmarshal(improvementSuggestionsJSON, &analytics.ImprovementSuggestions)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal improvement suggestions: %w", err)
		}
		
		analyticsList = append(analyticsList, &analytics)
	}
	
	return analyticsList, totalCount, nil
}

// CreatePredictiveMaintenanceAnalytics creates new predictive maintenance analytics
