var app = angular.module('huntApp', ['ngCookies']);

app.controller('huntCtrl', ['$scope', '$cookies', '$http', function ($scope, $cookies, $http) {
	$scope.refreshHunt = function() {
	    $http.get("/api/hunt?hunt_id=" + huntId).success(function (response) {
		    $scope.info = response;
		    $scope.signedIn = ($scope.info.Teams.CurrentTeam != "");
		    if ($scope.info.Teams.BadSignIn) {
			window.alert("Bad password");
			$scope.logout();
		    }
		});
	}

	$scope.login = function() {
	    $cookies.put("team_id", $scope.teamID);
	    $cookies.put("password", $scope.password);
	    $scope.teamID = "";
	    $scope.password = "";
	    $scope.refreshHunt();
	}

	$scope.logout = function() {
	    $cookies.remove("team_id");
	    $cookies.remove("password");
	    $scope.refreshHunt();
	}

	$scope.updatePuzzles = function() {
	    for (var i = 0; i < $scope.info.Puzzles.Puzzles.length; i++) {
		var p = $scope.info.Puzzles.Puzzles[i];
		$http.get("/api/updatepuzzle?hunt_id=" + huntId +
			  "&puzzleid=" + p.ID +
			  "&name=" + p.Name +
			  "&answer=" + p.Answer);
	    }
	}

	$scope.puzzleFormat = {
	    false: "Non-paper",
	    true: "Paper",
	}

	$scope.signedIn = false;
	$scope.teamID = "";
	$scope.password = "";
	$scope.refreshHunt();
    }]);