import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'src/router.dart';

void main() {
  runApp(const ProviderScope(child: ConnectSphereApp()));
}

class ConnectSphereApp extends ConsumerWidget {
  const ConnectSphereApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'ConnectSphere',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        useMaterial3: true,
        appBarTheme: const AppBarTheme(centerTitle: true),
      ),
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
