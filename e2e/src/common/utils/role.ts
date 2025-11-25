export function generateRoleAssignments(
  roles: string[],
  defaultAddress: string,
  overrides: { role: string; addresses: string[] }[],
): { role: string; addressWithRole: string }[] {
  const roleAssignments: { role: string; addressWithRole: string }[] = [];

  const overridesMap = new Map<string, string[]>();
  for (const override of overrides) {
    overridesMap.set(override.role, override.addresses);
  }

  const allRolesSet = new Set<string>(roles);
  for (const override of overrides) {
    allRolesSet.add(override.role);
  }

  for (const role of allRolesSet) {
    if (overridesMap.has(role)) {
      const addresses = overridesMap.get(role);

      if (addresses && addresses.length > 0) {
        for (const addressWithRole of addresses) {
          roleAssignments.push({ role, addressWithRole });
        }
      }
    } else {
      roleAssignments.push({ role, addressWithRole: defaultAddress });
    }
  }

  return roleAssignments;
}
