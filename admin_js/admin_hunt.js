var app = angular.module('huntApp', []);

app.factory('api', function($http) {
	var result = {};
	
	result.getURL = function (path) {
	    var result = "/api/" + path + "?hunt_id=" + huntId;
	    if (isAdmin) {
		result = "/admin" + result;
	    }
	    return result
	}

	result.getPuzzles = function() {
	    return $http.get(result.getURL("puzzles"));
	}

	result.getIngredients = function() {
	    return $http.get(result.getURL("ingredients"));
	}

	result.setIngredients = function(newIngredients) {
	    return $http.get(result.getURL("updateingredients") +
			     "&ingredients=" + encodeURIComponent(newIngredients));
	}

	result.getTeams = function() {
	    return $http.get(result.getURL("teams"));
	}

	result.addTeam = function(name, password, novice) {
	    return $http.get(result.getURL("addteam") + 
			     "&name=" + encodeURIComponent(name) +
			     "&password=" + encodeURIComponent(password) +
			     "&novice=" + novice);
	}

	result.deleteTeam = function(id) {
	    return $http.get(result.getURL("deleteteam") +
			     "&team_id=" + id);
	}

	result.getLeaderboard = function() {
	    return $http.get(result.getURL("leaderboard"));
	}

	result.getLeaderboardUpdate = function(id) {
	    return $http.get(result.getURL("leaderboardupdate") +
			     "&puzzleid=" + id);
	}

	return result;
    });

app.controller('leaderboardCtrl', function ($scope, api) {
	$scope.refresh = function() {
	    api.getLeaderboard().success(function (response) {
		    $scope.Leaderboard = response;
		});
	}

	$scope.updatePuzzles = function(id) {
	    api.getLeaderboardupdate(id).success(function (response) {
		    for (var i = 0; i < $scope.Leaderboard.Progress.length; i++) {
			if (id == $scope.Leaderboard.Progress[i].ID) {
			    $scope.Leaderboard.Progress[i].Updatable = response;
			}
		    }
		});
	}

	$scope.submitAnswer = function(pid, answer) {
	    // TODO(dneal)
	}

	$scope.puzzleFormat = {
	    false: "Non-paper",
	    true: "Paper",
	}

	$scope.refresh();
    });

app.controller('puzzlesCtrl', function ($scope, api) {
	$scope.refresh = function() {
	    api.getPuzzles().success(function (response) {
		    $scope.puzzles = response;
		    $scope.hasPuzzles = (response != null && response.length > 0);
		});
	}

	$scope.hasPuzzles = false;
	$scope.refresh();
    });

app.controller('ingredientsCtrl', function ($scope, api) {
	$scope.refresh = function() {
	    api.getIngredients().success(function (response) {
		    $scope.ingredients = response;
		});
	}

	$scope.updateIngredients = function() {
	    api.setIngredients($scope.newIngredients).success(function () {
		    $scope.newIngredients = "";
		    // TODO(dneal): Go through channel?
		    $scope.refresh();
		});
	}

	$scope.newIngredents = "";
	$scope.refresh();
    });

app.controller('teamsCtrl', function ($scope, api) {
	$scope.refresh = function() {
	    api.getTeams().success(function (response) {
		    $scope.teams = response;
		});
	}

	$scope.addTeam = function() {
	    api.addTeam($scope.newTeamName, $scope.newTeamPassword,
			$scope.newTeamNovice).success(function () {
				$scope.newTeamName = "";
				$scope.newTeamPassword = "";
				$scope.newTeamNovice = false;
				// TODO(dneal): Rely on channel?
				$scope.refresh();
			    });
	}

	$scope.deleteTeam = function(id) {
	    api.deleteTeam(id).success(function () {
		    // TODO(dneal): Rely on channel?
		    $scope.refresh();
		});
	}

	$scope.newTeamName = "";
	$scope.newTeamPassword = "";
	$scope.newTeamNovice = false;
	$scope.refresh();
    });

app.controller('adminHuntCtrl', function ($scope, $http) {
	$scope.refreshHunt = function() {
	    $http.get("/admin/api/hunt?hunt_id=" + huntId).success(function (response) {
		    $scope.hunt = response;
		    $scope.advanceable = ($scope.hunt.State != 5 && $scope.hunt.State != 7);
		});
	}

	
	$scope.advanceState = function() {
	    r = window.confirm("Are you sure you wish to advance the state? This cannot be undone.");
	    if (r) {
		$http.get("/admin/api/advancestate?hunt_id=" + huntId +
			  "&currentstate=" + $scope.hunt.State).success(function (response) {
				  $scope.refreshHunt();
			      });
	    }
	}

	$scope.stateTable = {
	    0: "Pre-launch",
	    1: "Novice ingredients released",
	    2: "Ingredients released",
	    3: "Solving",
	    4: "Surveying",
	    5: "Tallying results",
	    6: "Tallying done",
	    7: "Results released",
	}

	$scope.advanceable = false;
	$scope.refreshHunt();
    });