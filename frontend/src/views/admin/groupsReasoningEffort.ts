import type {
  GroupPlatform,
  ReasoningEffortMapping,
  ReasoningEffortMatchType,
  ReasoningEffortOverLimitPolicy,
} from "@/types";

const openAIReasoningEffortValues = [
  "minimal",
  "low",
  "medium",
  "high",
  "xhigh",
  "max",
] as const;

export const MAX_REASONING_EFFORT_MAPPINGS = 64;
export const MAX_REASONING_EFFORT_MODEL_LENGTH = 200;
const reasoningEffortMatchTypes: readonly ReasoningEffortMatchType[] = [
  "exact",
  "prefix",
  "suffix",
];

export const reasoningEffortOverLimitDowngrade: ReasoningEffortOverLimitPolicy =
  "downgrade";
export const reasoningEffortOverLimitDeny: ReasoningEffortOverLimitPolicy =
  "deny";

const reasoningEffortValuesForPlatform = (
  platform: GroupPlatform,
): readonly string[] =>
  supportsReasoningEffortPolicyPlatform(platform)
    ? openAIReasoningEffortValues
    : [];

export function supportsReasoningEffortPolicyPlatform(
  platform: GroupPlatform,
): boolean {
  return platform === "openai" || platform === "composite";
}

export function reasoningEffortOptionsForPlatform(platform: GroupPlatform) {
  return reasoningEffortValuesForPlatform(platform).map((value) => ({
    value,
    label: value,
  }));
}

export function normalizeReasoningEffortForPlatform(
  platform: GroupPlatform,
  value: string | null | undefined,
): string {
  const normalized = value?.trim().toLowerCase() ?? "";
  return reasoningEffortValuesForPlatform(platform).some(
    (allowed) => allowed === normalized,
  )
    ? normalized
    : "";
}

export function normalizeReasoningEffortMatchType(
  value: string | null | undefined,
): ReasoningEffortMatchType | "" {
  const normalized = value?.trim().toLowerCase() ?? "";
  return reasoningEffortMatchTypes.includes(normalized as ReasoningEffortMatchType)
    ? (normalized as ReasoningEffortMatchType)
    : "";
}

export function normalizeReasoningEffortOverLimit(
  value: string | null | undefined,
): ReasoningEffortOverLimitPolicy {
  return value?.trim().toLowerCase() === reasoningEffortOverLimitDeny
    ? reasoningEffortOverLimitDeny
    : reasoningEffortOverLimitDowngrade;
}

export interface ReasoningEffortMappingRow extends ReasoningEffortMapping {
  id: string;
}

export type ReasoningEffortMappingErrorCode =
  | "fromRequired"
  | "toRequired"
  | "duplicateFrom"
  | "unsupportedFrom"
  | "unsupportedTo"
  | "unsupportedMatchType"
  | "modelRequired"
  | "modelTooLong"
  | "mappingLimit";

export type ReasoningEffortMappingErrors = Record<
  string,
  Partial<
    Record<
      "from" | "to" | "match_type" | "model" | "limit",
      ReasoningEffortMappingErrorCode
    >
  >
>;

let nextMappingRowID = 0;

export function createReasoningEffortMappingRow(
  mapping: Partial<ReasoningEffortMapping> = {},
): ReasoningEffortMappingRow {
  nextMappingRowID += 1;
  return {
    id: `reasoning-effort-mapping-${nextMappingRowID}`,
    from: mapping.from ?? "",
    to: mapping.to ?? "",
    match_type: normalizeReasoningEffortMatchType(mapping.match_type) || undefined,
    model: mapping.model?.trim() ?? "",
  };
}

export function reasoningEffortMappingsToRows(
  mappings?: ReasoningEffortMapping[] | null,
  platform: GroupPlatform = "openai",
): ReasoningEffortMappingRow[] {
  return (mappings ?? []).flatMap((mapping) => {
    const from = normalizeReasoningEffortForPlatform(platform, mapping.from);
    const to = normalizeReasoningEffortForPlatform(platform, mapping.to);
    const model = mapping.model?.trim() ?? "";
    const matchType = normalizeReasoningEffortMatchType(mapping.match_type);
    return from && to
      ? [createReasoningEffortMappingRow({ from, to, model, match_type: matchType || undefined })]
      : [];
  });
}

export function reasoningEffortMappingsToAPI(
  rows: ReasoningEffortMappingRow[],
): ReasoningEffortMapping[] {
  return rows.map((row) => ({
    from: row.from.trim(),
    to: row.to.trim(),
    ...(row.model?.trim()
      ? {
          model: row.model.trim(),
          match_type: normalizeReasoningEffortMatchType(row.match_type) || "exact",
        }
      : {}),
  }));
}

export function validateReasoningEffortMappings(
  rows: ReasoningEffortMappingRow[],
  platform: GroupPlatform = "openai",
): ReasoningEffortMappingErrors {
  const errors: ReasoningEffortMappingErrors = {};
  const sourceRows = new Map<string, ReasoningEffortMappingRow[]>();

  if (rows.length > MAX_REASONING_EFFORT_MAPPINGS) {
    rows.forEach((row) => {
      errors[row.id] = { ...errors[row.id], limit: "mappingLimit" };
    });
  }

  rows.forEach((row) => {
    const from = row.from.trim();
    const to = row.to.trim();
    const model = row.model?.trim() ?? "";
    const rawMatchType = row.match_type?.trim().toLowerCase() ?? "";
    const matchType = normalizeReasoningEffortMatchType(row.match_type);
    if (rawMatchType && !matchType) {
      errors[row.id] = { ...errors[row.id], match_type: "unsupportedMatchType" };
    }
    if (matchType && !model) {
      errors[row.id] = { ...errors[row.id], model: "modelRequired" };
    }
    if (model.length > MAX_REASONING_EFFORT_MODEL_LENGTH) {
      errors[row.id] = { ...errors[row.id], model: "modelTooLong" };
    }
    const scopeKey = `${model.toLowerCase()}\0${model ? matchType || "exact" : ""}`;
    if (!from) {
      errors[row.id] = { ...errors[row.id], from: "fromRequired" };
    } else if (!normalizeReasoningEffortForPlatform(platform, from)) {
      errors[row.id] = { ...errors[row.id], from: "unsupportedFrom" };
    } else {
      const key = `${scopeKey}\0${from.toLowerCase()}`;
      sourceRows.set(key, [...(sourceRows.get(key) ?? []), row]);
    }
    if (!to) {
      errors[row.id] = { ...errors[row.id], to: "toRequired" };
    } else if (!normalizeReasoningEffortForPlatform(platform, to)) {
      errors[row.id] = { ...errors[row.id], to: "unsupportedTo" };
    }
  });

  sourceRows.forEach((duplicateRows) => {
    if (duplicateRows.length < 2) return;
    duplicateRows.forEach((row) => {
      errors[row.id] = { ...errors[row.id], from: "duplicateFrom" };
    });
  });

  return errors;
}
