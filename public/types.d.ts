// Type definitions for Custom Elements
// These augment the DOM types for custom element lifecycle callbacks

/** Interface for Custom Elements with lifecycle callbacks */
interface CustomElement extends HTMLElement {
    /** Called when element is added to the DOM */
    connectedCallback?(): void

    /** Called when element is removed from the DOM */
    disconnectedCallback?(): void

    /** Called when element is moved to a new document */
    adoptedCallback?(): void

    /** Called when an observed attribute changes */
    attributeChangedCallback?(name: string, oldValue: string | null, newValue: string | null): void
}

interface CustomElementConstructor {
    new (...params: any[]): CustomElement
    readonly observedAttributes?: string[]
}

interface CustomElementRegistry {
    define(
        name: string,
        constructor: CustomElementConstructor,
        options?: ElementDefinitionOptions,
    ): void
    get(name: string): CustomElementConstructor | undefined
    getName(constructor: CustomElementConstructor): string | null
    whenDefined(name: string): Promise<CustomElementConstructor>
}
