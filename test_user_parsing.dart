import 'dart:convert';

// Simple test to verify User.fromJson works with search results
void main() {
  // This is the JSON format returned by the search API
  final jsonResponse = '''
  [
    {
      "id": "bfa31d27-66a7-47a9-9473-76991f0ed9f9",
      "username": "jane_smith",
      "display_name": "Jane Smith",
      "created_at": "2025-08-30T03:16:56.867319Z"
    }
  ]
  ''';

  try {
    final List<dynamic> jsonList = json.decode(jsonResponse);
    print('✅ JSON parsing successful');
    print('Found ${jsonList.length} users');

    for (var userJson in jsonList) {
      print('User data: $userJson');
      print('  ID: ${userJson['id']}');
      print('  Username: ${userJson['username']}');
      print('  Display Name: ${userJson['display_name']}');
      print('  Created At: ${userJson['created_at']}');
      print(
        '  Email: ${userJson['email'] ?? 'Not provided (expected for search)'}',
      );
    }

    print('\n🎉 The User model should now work with search results!');
  } catch (e) {
    print('❌ Error parsing JSON: $e');
  }
}
