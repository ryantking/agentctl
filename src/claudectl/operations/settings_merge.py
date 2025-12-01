"""Settings merge utilities for smart JSON merging."""

from __future__ import annotations


def merge_settings_smart(base: dict, overlay: dict) -> dict:
    """Deep merge settings with intelligent array handling.

    Strategy:
    - Nested dicts: Recursive merge
    - Arrays: Union (deduplicate simple types)
    - Scalars: Overlay takes precedence

    Args:
        base: Existing settings
        overlay: New settings to merge in

    Returns:
        Merged settings dictionary
    """
    result = base.copy()

    for key, value in overlay.items():
        if key not in result:
            # New key - add it
            result[key] = value
        elif isinstance(value, dict) and isinstance(result[key], dict):
            # Both dicts - recursive merge
            result[key] = merge_settings_smart(result[key], value)
        elif isinstance(value, list) and isinstance(result[key], list):
            # Both lists - merge with deduplication
            result[key] = _merge_lists(result[key], value)
        else:
            # Scalar or type mismatch - overlay wins
            result[key] = value

    return result


def _merge_lists(base: list, overlay: list) -> list:
    """Merge two lists with deduplication for simple types."""
    result = base.copy()

    for item in overlay:
        # For simple types (str, int, bool), deduplicate
        if isinstance(item, (str, int, bool, float, type(None))):
            if item not in result:
                result.append(item)
        else:
            # For complex types (dict, list), just append
            result.append(item)

    return result
