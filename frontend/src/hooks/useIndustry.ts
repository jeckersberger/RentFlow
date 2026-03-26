import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../services/api'
import { INDUSTRY_PROFILES, DEFAULT_LABELS } from '../config/industryProfiles'
import type { IndustryProfile } from '../config/industryProfiles'

/**
 * Hook to access the current tenant's industry profile.
 *
 * Returns the resolved profile, a label() helper that returns
 * industry-specific labels, and a hasFeature() helper.
 */
export function useIndustry() {
  const queryClient = useQueryClient()

  const { data: configData } = useQuery({
    queryKey: ['config', 'industry_profile'],
    queryFn: () => configApi.get('industry_profile'),
    staleTime: 1000 * 60 * 60, // 1 hour cache
  })

  // The config endpoint returns { data: "event_tech" } or similar
  const profileId: string =
    (typeof configData?.data === 'string' ? configData.data : configData?.data) || 'event_tech'

  const profile: IndustryProfile =
    INDUSTRY_PROFILES[profileId] || INDUSTRY_PROFILES['event_tech']

  const setIndustryMutation = useMutation({
    mutationFn: (newProfileId: string) => configApi.set('industry_profile', newProfileId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'industry_profile'] })
    },
  })

  return {
    profileId,
    profile,
    /** Get an industry-specific label for a given key */
    label: (key: string): string =>
      profile.labels[key] || DEFAULT_LABELS[key] || key,
    /** Check if a feature is enabled for this industry */
    hasFeature: (feature: string): boolean =>
      profile.features[feature] || false,
    /** The categories for this industry */
    categories: profile.categories,
    /** The scanner home cards for this industry */
    scannerHomeCards: profile.scannerHomeCards,
    /** Change the industry profile */
    setIndustry: setIndustryMutation.mutate,
    isSettingIndustry: setIndustryMutation.isPending,
  }
}
