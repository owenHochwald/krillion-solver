import SwiftUI

@main
struct KrillionBarApp: App {
    @StateObject private var store = AnswerStore()

    var body: some Scene {
        MenuBarExtra {
            AnswerPanel(store: store)
                .frame(width: 350)
                .frame(maxHeight: 610)
                .onAppear { store.loadIfNeeded() }
        } label: {
            Image("MenuBarShrimp")
                .renderingMode(.template)
                .accessibilityLabel("Krillion Solver")
        }
        .menuBarExtraStyle(.window)
    }
}
