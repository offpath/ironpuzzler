var app = angular.module('huntApp', []);

app.factory('api', function($http) {
	var result = {};
	var listeners = [];
	var channel;
	var socket;
	
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

	result.getConsole = function(id) {
	    return $http.get(result.getURL("console"));
	}

	result.addListener = function(listener) {
	    listeners.push(listener);
	}

	result.onMessage = function(message) {
	    var j = JSON.parse(message.data);
	    if (j.K == "refresh") {
		for (var i = 0; i < listeners.length; i++) {
		    if (listeners[i].hasOwnProperty("refresh")) {
			listeners[i].refresh();
		    }
		}
	    } else {
		for (var i = 0; i < listeners.length; i++) {
		    if (listeners[i].hasOwnProperty("onMessage")) {
			listeners[i].onMessage(j);
		    }
		}
	    }
	}

	result.openChannel = function() {
	    $http.get(result.getURL("channel")).success(function (response) {
		    channel = new goog.appengine.Channel(response)
		    socket = channel.open();
		    socket.onmessage = result.onMessage;
		});
	}

	result.openChannel();
	return result;
    });

app.controller('consoleCtrl', function($scope, api) {
	$scope.onMessage = function(message) {
	    if (message.K == "consoleupdate") {
		$scope.$apply(function() {
			$scope.Lines.unshift(message.V);
		    });
	    }
	}

	$scope.Lines = [];
	api.getConsole().success(function (response) {
		$scope.Lines = response;
	    });
	api.addListener($scope);
    });

app.controller('leaderboardCtrl', function ($scope, api) {
	$scope.refresh = function() {
	    api.getLeaderboard().success(function (response) {
		    $scope.Leaderboard = response;
		});
	}

	$scope.updatePuzzle = function(id) {
	    api.getLeaderboardUpdate(id).success(function (response) {
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

	$scope.onMessage = function(message) {
	    if (message.K == "leaderboardupdate") {
		$scope.updatePuzzle(message.V);
	    }
	}

	api.addListener($scope);
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