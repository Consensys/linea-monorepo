export interface LineaRollupV7ReinitializationData {
  pauseTypeRoles: { pauseType: number; role: string }[];
  unpauseTypeRoles: { pauseType: number; role: string }[];
  roleAddresses: { addressWithRole: string; role: string }[];
  yieldManager: string;
}
