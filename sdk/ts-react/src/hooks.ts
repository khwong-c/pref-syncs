import { useContext } from 'react';
import { PrefSyncContext } from './provider';
import type { PrefSyncContextValue } from './types';

/**
 * Hook for accessing the active Preference Sync client and context state.
 * Throws an error if used outside of a PrefSyncProvider.
 */
export const usePrefSync = (): PrefSyncContextValue => {
  const context = useContext(PrefSyncContext);
  if (!context) {
    throw new Error('usePrefSync must be used within a PrefSyncProvider.');
  }
  return context;
};
