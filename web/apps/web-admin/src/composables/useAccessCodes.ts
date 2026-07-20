import { computed } from 'vue';

import { useAccessStore } from '@vben/stores';

type AccessStoreWithCodes = {
  accessCodes?: string[];
};

export function useAccessCodes() {
  const accessStore = useAccessStore() as unknown as AccessStoreWithCodes;
  const codeSet = computed(() => new Set(accessStore.accessCodes ?? []));

  function hasAccessCode(code: string) {
    return codeSet.value.has(code);
  }

  function hasAnyAccessCode(codes: string[]) {
    return codes.some((code) => hasAccessCode(code));
  }

  return {
    hasAccessCode,
    hasAnyAccessCode,
  };
}
