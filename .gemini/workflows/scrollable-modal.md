---
description: Rule for scrollable modals/popups when content exceeds viewport height
---

# Scrollable Modal/Popup Rule

## Rule
**Every modal or popup in this project MUST be scrollable if its content exceeds the viewport height.**

## Implementation
On the modal's **inner content div** (not the backdrop overlay), always add:

```
max-h-[90vh] overflow-y-auto
```

### Example Structure
```html
<!-- Backdrop (outer) — NO scroll here -->
<div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm">
    <!-- Content (inner) — ADD scroll here -->
    <div class="glass-card bg-white p-6 rounded-2xl w-full max-w-sm max-h-[90vh] overflow-y-auto relative z-10 shadow-2xl">
        <!-- Form content here -->
    </div>
</div>
```

## Checklist
When creating or modifying any modal/popup:
1. ✅ Add `max-h-[90vh]` to the inner content div
2. ✅ Add `overflow-y-auto` to the inner content div  
3. ❌ Do NOT add scroll to the backdrop/overlay div
4. ❌ Do NOT add scroll to internal lists separately (let the whole form scroll)
